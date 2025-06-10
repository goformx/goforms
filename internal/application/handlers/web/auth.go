package web

import (
	"net/http"
	"time"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	deps HandlerDeps
}

const (
	// SessionDuration is the duration for which a session remains valid
	SessionDuration = 24 * time.Hour
	// MinPasswordLength is the minimum required length for passwords
	MinPasswordLength = 8
)

// NewAuthHandler creates a new auth handler
func NewAuthHandler(deps HandlerDeps) (*AuthHandler, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return &AuthHandler{
		deps: deps,
	}, nil
}

// Register registers the auth handler routes
func (h *AuthHandler) Register(e *echo.Echo) {
	// Public routes
	e.GET("/login", h.Login)
	e.POST("/login", h.LoginPost)
	e.GET("/signup", h.Signup)
	e.POST("/signup", h.SignupPost)
	e.POST("/logout", h.Logout)

	// Auth group for protected routes
	auth := e.Group("/auth")
	// Validation schema endpoints
	auth.GET("/api/validation/login", h.handleLoginValidation)
	auth.GET("/api/validation/signup", h.handleSignupValidation)
}

// Login handles the login page request
func (h *AuthHandler) Login(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, "Login")
	return h.deps.Renderer.Render(c, pages.Login(data))
}

// LoginPost handles the login form submission
func (h *AuthHandler) LoginPost(c echo.Context) error {
	// TODO: Implement login logic
	return response.WebErrorResponse(c, h.deps.Renderer, http.StatusNotImplemented, "Login not implemented")
}

// Signup handles the signup page request
func (h *AuthHandler) Signup(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, "Sign Up")
	return h.deps.Renderer.Render(c, pages.Signup(data))
}

// SignupPost handles the signup form submission
func (h *AuthHandler) SignupPost(c echo.Context) error {
	// TODO: Implement signup logic
	return response.WebErrorResponse(c, h.deps.Renderer, http.StatusNotImplemented, "Signup not implemented")
}

// Logout handles the logout request
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get session cookie
	cookie, err := c.Cookie(h.deps.SessionManager.GetCookieName())
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Delete session
	h.deps.SessionManager.DeleteSession(cookie.Value)

	// Clear session cookie
	h.deps.SessionManager.ClearSessionCookie(c)

	return c.Redirect(http.StatusSeeOther, "/login")
}

// handleLoginValidation handles the login form validation schema request
func (h *AuthHandler) handleLoginValidation(c echo.Context) error {
	schema := map[string]any{
		"email": map[string]any{
			"type":    "email",
			"message": "Please enter a valid email address",
		},
		"password": map[string]any{
			"type":    "password",
			"min":     MinPasswordLength, // Minimum password length requirement
			"message": "Password must be at least 8 characters long",
		},
	}
	return c.JSON(http.StatusOK, schema)
}

// handleSignupValidation returns the validation schema for the signup form
func (h *AuthHandler) handleSignupValidation(c echo.Context) error {
	schema := map[string]any{
		"first_name": map[string]any{
			"type":    "string",
			"min":     1,
			"message": "First name is required",
		},
		"last_name": map[string]any{
			"type":    "string",
			"min":     1,
			"message": "Last name is required",
		},
		"email": map[string]any{
			"type":    "email",
			"message": "Please enter a valid email address",
		},
		"password": map[string]any{
			"type":    "password",
			"min":     MinPasswordLength, // Minimum password length requirement
			"message": "Password must be at least 8 characters long",
		},
		"confirm_password": map[string]any{
			"type":       "match",
			"matchField": "password",
			"message":    "Passwords must match",
		},
	}
	return c.JSON(http.StatusOK, schema)
}
