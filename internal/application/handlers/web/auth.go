package web

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/request"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/web"
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
	AssetManager    *web.AssetManager
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
	assetManager *web.AssetManager,
) (*AuthHandler, error) {
	if base == nil || authMiddleware == nil || requestUtils == nil ||
		schemaGenerator == nil || requestParser == nil || responseBuilder == nil ||
		authService == nil || sanitizer == nil || assetManager == nil {
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
		AssetManager:    assetManager,
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
	return c.JSON(constants.StatusOK, map[string]string{
		"message": "Test endpoint working",
		"status":  "success",
	})
}

// Login handles GET /login - displays the login form
func (h *AuthHandler) Login(c echo.Context) error {
	// Force CSRF token generation
	_ = c.Get("csrf")
	data := view.BuildPageData(h.Config, h.AssetManager, c, "Login")
	if mwcontext.IsAuthenticated(c) {
		if err := c.Redirect(constants.StatusSeeOther, constants.PathDashboard); err != nil {
			return fmt.Errorf("redirect to dashboard: %w", err)
		}
		return nil
	}
	if err := h.Renderer.Render(c, pages.Login(data)); err != nil {
		return fmt.Errorf("render login page: %w", err)
	}
	return nil
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
	email, password, err := h.RequestParser.ParseLogin(c)
	if err != nil {
		h.Logger.Error("failed to parse login request", "error", err)
		if c.Request().Header.Get(constants.HeaderXRequestedWith) == XMLHttpRequestHeader {
			return h.ResponseBuilder.AJAXError(
				c, constants.StatusBadRequest, constants.ErrMsgInvalidRequest,
			)
		}
		data := view.BuildPageData(h.Config, h.AssetManager, c, "Login")
		return h.ResponseBuilder.HTMLFormError(c, "login", &data, constants.ErrMsgInvalidRequest)
	}

	email = h.Sanitizer.Email(email)

	_, sessionID, err := h.AuthService.Login(c.Request().Context(), email, password, c.Request().UserAgent())
	if err != nil {
		h.Logger.Error("login failed", "error", err)
		if c.Request().Header.Get(constants.HeaderXRequestedWith) == XMLHttpRequestHeader {
			return h.ResponseBuilder.AJAXError(
				c, constants.StatusUnauthorized, constants.ErrMsgInvalidCredentials,
			)
		}
		data := view.BuildPageData(h.Config, h.AssetManager, c, "Login")
		return h.ResponseBuilder.HTMLFormError(c, "login", &data, constants.ErrMsgInvalidCredentials)
	}

	h.SessionManager.SetSessionCookie(c, sessionID)

	if c.Request().Header.Get(constants.HeaderXRequestedWith) == XMLHttpRequestHeader {
		return c.JSON(constants.StatusOK, map[string]string{
			"redirect": constants.PathDashboard,
		})
	}
	return h.ResponseBuilder.Redirect(c, constants.PathDashboard)
}

// Signup handles GET /signup - displays the signup form
func (h *AuthHandler) Signup(c echo.Context) error {
	// Force CSRF token generation
	_ = c.Get("csrf")
	data := view.BuildPageData(h.Config, h.AssetManager, c, "Sign Up")
	if mwcontext.IsAuthenticated(c) {
		if err := c.Redirect(constants.StatusSeeOther, constants.PathDashboard); err != nil {
			return fmt.Errorf("redirect to dashboard: %w", err)
		}
		return nil
	}
	if err := h.Renderer.Render(c, pages.Signup(data)); err != nil {
		return fmt.Errorf("render signup page: %w", err)
	}
	return nil
}

// SignupPost handles the signup form submission
func (h *AuthHandler) SignupPost(c echo.Context) error {
	signup, err := h.RequestParser.ParseSignup(c)
	if err != nil {
		h.Logger.Error("failed to parse signup request", "error", err)
		if c.Request().Header.Get(constants.HeaderXRequestedWith) == XMLHttpRequestHeader {
			return h.ResponseBuilder.AJAXError(
				c,
				constants.StatusBadRequest,
				constants.ErrMsgInvalidRequest,
			)
		}
		data := view.BuildPageData(h.Config, h.AssetManager, c, "Sign Up")
		return h.ResponseBuilder.HTMLFormError(c, "signup", &data, constants.ErrMsgInvalidRequest)
	}

	signup.Email = h.Sanitizer.Email(signup.Email)

	_, sessionID, err := h.AuthService.Signup(c.Request().Context(), signup, c.Request().UserAgent())
	if err != nil {
		h.Logger.Error("signup failed", "error", err)
		if c.Request().Header.Get(constants.HeaderXRequestedWith) == XMLHttpRequestHeader {
			return h.ResponseBuilder.AJAXError(
				c,
				constants.StatusBadRequest,
				"Unable to create account. Please try again.",
			)
		}
		data := view.BuildPageData(h.Config, h.AssetManager, c, "Sign Up")
		return h.ResponseBuilder.HTMLFormError(
			c,
			"signup",
			&data,
			"Unable to create account. Please try again.",
		)
	}

	h.SessionManager.SetSessionCookie(c, sessionID)

	if c.Request().Header.Get(constants.HeaderXRequestedWith) == XMLHttpRequestHeader {
		return c.JSON(constants.StatusOK, map[string]string{
			"message":  constants.MsgSignupSuccess,
			"redirect": constants.PathDashboard,
		})
	}
	return h.ResponseBuilder.Redirect(c, constants.PathDashboard)
}

// Logout handles POST /logout - processes the logout request
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get session cookie
	cookie, err := c.Cookie(h.SessionManager.GetCookieName())
	if err != nil {
		return fmt.Errorf("redirect to login: %w", c.Redirect(constants.StatusSeeOther, constants.PathLogin))
	}

	// Delete session
	h.SessionManager.DeleteSession(cookie.Value)

	// Clear session cookie
	h.SessionManager.ClearSessionCookie(c)

	return fmt.Errorf("redirect to login: %w", c.Redirect(constants.StatusSeeOther, constants.PathLogin))
}

// LoginValidation handles the login form validation schema request
func (h *AuthHandler) LoginValidation(c echo.Context) error {
	// Generate schema using the validation package
	schema := h.SchemaGenerator.GenerateLoginSchema()

	return c.JSON(constants.StatusOK, schema)
}

// SignupValidation returns the validation schema for the signup form
func (h *AuthHandler) SignupValidation(c echo.Context) error {
	// Generate schema using the validation package
	schema := h.SchemaGenerator.GenerateSignupSchema()

	return c.JSON(constants.StatusOK, schema)
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
