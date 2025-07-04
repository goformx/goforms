package web

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/request"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	*BaseHandler
	AuthMiddleware  *auth.Middleware
	RequestUtils    *request.Utils
	SchemaGenerator *validation.SchemaGenerator
	RequestParser   *AuthRequestParser
	ResponseBuilder *AuthResponseBuilder
	AuthService     *AuthService
	Sanitizer       sanitization.ServiceInterface
}

const (
	// SessionDuration is the duration for which a session remains valid
	SessionDuration = 24 * time.Hour
	// MinPasswordLength is the minimum required length for passwords
	MinPasswordLength = 8
	// XMLHttpRequestHeader is the standard header value for AJAX requests
	XMLHttpRequestHeader = "XMLHttpRequest"
)

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	base *BaseHandler,
	authMiddleware *auth.Middleware,
	requestUtils *request.Utils,
	schemaGenerator *validation.SchemaGenerator,
	requestParser *AuthRequestParser,
	responseBuilder *AuthResponseBuilder,
	authService *AuthService,
	sanitizer sanitization.ServiceInterface,
) (*AuthHandler, error) {
	if base == nil || authMiddleware == nil || requestUtils == nil ||
		schemaGenerator == nil || requestParser == nil || responseBuilder == nil ||
		authService == nil || sanitizer == nil {
		return nil, errors.New("missing required dependencies for AuthHandler")
	}

	return &AuthHandler{
		BaseHandler:     base,
		AuthMiddleware:  authMiddleware,
		RequestUtils:    requestUtils,
		SchemaGenerator: schemaGenerator,
		RequestParser:   requestParser,
		ResponseBuilder: responseBuilder,
		AuthService:     authService,
		Sanitizer:       sanitizer,
	}, nil
}

// Register registers the auth handler routes
// Note: Routes are actually registered by RegisterHandlers in module.go
func (h *AuthHandler) Register(e *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
	// Add a simple test endpoint to verify JSON responses work
	api := e.Group(constants.PathAPIv1)
	api.GET("/test", h.TestEndpoint)
}

// TestEndpoint is a simple test endpoint to verify JSON responses work
func (h *AuthHandler) TestEndpoint(c echo.Context) error {
	return response.Success(c, map[string]string{
		"message": "Test endpoint working",
		"status":  "success",
	})
}

// Login handles GET /login - displays the login form
func (h *AuthHandler) Login(c echo.Context) error {
	return h.renderAuthPage(c, "Login", pages.Login)
}

// LoginPost handles POST /login - processes the login form
//
// This handler:
// 1. Validates user credentials
// 2. Creates a new session on success
// 3. Sets session cookie
// 4. Returns appropriate response based on request type:
//   - JSON response for API requests
//   - HTML response with error for regular requests
//   - Redirect to dashboard on success
func (h *AuthHandler) LoginPost(c echo.Context) error {
	return h.handleAuthSubmission(c, h.processLogin, "Login")
}

// Signup handles GET /signup - displays the signup form
func (h *AuthHandler) Signup(c echo.Context) error {
	return h.renderAuthPage(c, "Sign Up", pages.Signup)
}

// SignupPost handles the signup form submission
func (h *AuthHandler) SignupPost(c echo.Context) error {
	return h.handleAuthSubmission(c, h.processSignup, "Sign Up")
}

// Logout handles POST /logout - processes the logout request
func (h *AuthHandler) Logout(c echo.Context) error {
	if err := h.clearUserSession(c); err != nil {
		h.Logger.Error("failed to clear session", "error", err)
	}

	return h.redirectToLogin(c)
}

// LoginValidation handles the login form validation schema request
func (h *AuthHandler) LoginValidation(c echo.Context) error {
	schema := h.SchemaGenerator.GenerateLoginSchema()

	return response.Success(c, schema)
}

// SignupValidation returns the validation schema for the signup form
func (h *AuthHandler) SignupValidation(c echo.Context) error {
	schema := h.SchemaGenerator.GenerateSignupSchema()

	return response.Success(c, schema)
}

// Start initializes the auth handler.
// This is called during application startup.
func (h *AuthHandler) Start(_ context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the auth handler.
// This is called during application shutdown.
func (h *AuthHandler) Stop(_ context.Context) error {
	return nil // No cleanup needed
}

// Helper methods for DRY and SRP compliance

// renderAuthPage handles the common pattern for rendering auth pages
func (h *AuthHandler) renderAuthPage(
	c echo.Context,
	title string,
	templateFunc func(view.PageData) templ.Component,
) error {
	// Force CSRF token generation
	_ = c.Get("csrf")

	data := h.NewPageData(c, title)

	if mwcontext.IsAuthenticated(c) {
		return h.redirectToDashboard(c)
	}

	return fmt.Errorf("render auth page: %w", h.Renderer.Render(c, templateFunc(*data)))
}

// handleAuthSubmission handles the common pattern for auth form submissions
func (h *AuthHandler) handleAuthSubmission(c echo.Context, processor func(echo.Context) error, pageTitle string) error {
	if err := processor(c); err != nil {
		return h.handleAuthError(c, err, pageTitle)
	}

	return h.handleAuthSuccess(c)
}

// processLogin handles login-specific processing
func (h *AuthHandler) processLogin(c echo.Context) error {
	email, password, err := h.RequestParser.ParseLogin(c)
	if err != nil {
		h.Logger.Error("failed to parse login request", "error", err)

		return fmt.Errorf("parse login: %w", err)
	}

	email = h.Sanitizer.Email(email)

	_, sessionID, err := h.AuthService.Login(c.Request().Context(), email, password, c.Request().UserAgent())
	if err != nil {
		h.Logger.Error("login failed", "error", err)

		return fmt.Errorf("login: %w", err)
	}

	h.SessionManager.SetSessionCookie(c, sessionID)

	return nil
}

// processSignup handles signup-specific processing
func (h *AuthHandler) processSignup(c echo.Context) error {
	signup, err := h.RequestParser.ParseSignup(c)
	if err != nil {
		h.Logger.Error("failed to parse signup request", "error", err)

		return fmt.Errorf("parse signup: %w", err)
	}

	signup.Email = h.Sanitizer.Email(signup.Email)

	_, sessionID, err := h.AuthService.Signup(c.Request().Context(), signup, c.Request().UserAgent())
	if err != nil {
		h.Logger.Error("signup failed", "error", err)

		return fmt.Errorf("signup: %w", err)
	}

	h.SessionManager.SetSessionCookie(c, sessionID)

	return nil
}

// handleAuthError handles authentication errors with appropriate response format
func (h *AuthHandler) handleAuthError(c echo.Context, err error, pageTitle string) error {
	if h.isAJAXRequest(c) {
		return h.ResponseBuilder.AJAXError(c, constants.StatusBadRequest, err.Error())
	}

	data := h.NewPageData(c, pageTitle)

	return h.ResponseBuilder.HTMLFormError(c, pageTitle, data, err.Error())
}

// handleAuthSuccess handles successful authentication with appropriate response format
func (h *AuthHandler) handleAuthSuccess(c echo.Context) error {
	if h.isAJAXRequest(c) {
		return response.Success(c, map[string]string{
			"redirect": constants.PathDashboard,
		})
	}

	return h.ResponseBuilder.Redirect(c, constants.PathDashboard)
}

// clearUserSession clears the user's session
func (h *AuthHandler) clearUserSession(c echo.Context) error {
	cookie, err := c.Cookie(h.SessionManager.GetCookieName())
	if err != nil {
		return fmt.Errorf("get session cookie: %w", err)
	}

	h.SessionManager.DeleteSession(cookie.Value)
	h.SessionManager.ClearSessionCookie(c)

	return nil
}

// redirectToDashboard redirects to the dashboard
func (h *AuthHandler) redirectToDashboard(c echo.Context) error {
	return c.Redirect(constants.StatusSeeOther, constants.PathDashboard)
}

// redirectToLogin redirects to the login page
func (h *AuthHandler) redirectToLogin(c echo.Context) error {
	return c.Redirect(constants.StatusSeeOther, constants.PathLogin)
}

// isAJAXRequest checks if the request is an AJAX request
func (h *AuthHandler) isAJAXRequest(c echo.Context) bool {
	return c.Request().Header.Get(constants.HeaderXRequestedWith) == XMLHttpRequestHeader
}
