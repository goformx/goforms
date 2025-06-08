package web

import (
	"net/http"
	"time"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	HandlerDeps
}

const (
	// SessionDuration is the duration for which a session remains valid
	SessionDuration = 24 * time.Hour
	// MinPasswordLength is the minimum required length for passwords
	MinPasswordLength = 8
)

// NewAuthHandler creates a new auth handler using HandlerDeps
func NewAuthHandler(deps HandlerDeps) (*AuthHandler, error) {
	if err := deps.Validate(
		"BaseHandler",
		"UserService",
		"SessionManager",
		"Renderer",
		"MiddlewareManager",
		"Config",
		"Logger",
	); err != nil {
		return nil, err
	}
	return &AuthHandler{HandlerDeps: deps}, nil
}

// Register registers the auth routes
func (h *AuthHandler) Register(e *echo.Echo) {
	e.GET("/login", h.showLoginPage)
	e.POST("/login", h.handleLogin)
	e.GET("/signup", h.showSignupPage)
	e.POST("/signup", h.handleSignup)
	e.POST("/logout", h.handleLogout)

	// Validation schema endpoints
	e.GET("/api/validation/login", h.handleLoginValidation)
	e.GET("/api/validation/signup", h.handleSignupValidation)
}

// showLoginPage renders the login page
func (h *AuthHandler) showLoginPage(c echo.Context) error {
	data := shared.BuildPageData(h.Config, "Login")
	return h.Renderer.Render(c, pages.Login(data))
}

// handleLogin processes the login request
func (h *AuthHandler) handleLogin(c echo.Context) error {
	email := c.FormValue("email")
	email = strings.ReplaceAll(email, "\n", "")
	email = strings.ReplaceAll(email, "\r", "")
	password := c.FormValue("password")

	h.Logger.Debug("login attempt",
		logging.StringField("email", email),
		logging.StringField("path", c.Request().URL.Path),
		logging.StringField("method", c.Request().Method),
	)

	// Authenticate user
	userData, err := h.UserService.Authenticate(c.Request().Context(), email, password)
	if err != nil {
		h.Logger.Error("failed to authenticate user", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}

	h.Logger.Debug("user authenticated",
		logging.UintField("user_id", userData.ID),
		logging.StringField("email", userData.Email),
		logging.StringField("role", userData.Role),
	)

	// Create session and set session cookie via SessionManager
	sessionID, err := h.SessionManager.CreateSession(userData.ID, userData.Email, userData.Role)
	if err != nil {
		h.Logger.Error("failed to create session", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create session")
	}

	h.Logger.Debug("session created",
		logging.StringField("session_id", sessionID),
		logging.UintField("user_id", userData.ID),
	)

	h.SessionManager.SetSessionCookie(c, sessionID)

	h.Logger.Debug("redirecting to dashboard",
		logging.StringField("session_id", sessionID),
		logging.UintField("user_id", userData.ID),
	)

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// showSignupPage renders the signup page
func (h *AuthHandler) showSignupPage(c echo.Context) error {
	data := shared.BuildPageData(h.Config, "Sign Up")
	return h.Renderer.Render(c, pages.Signup(data))
}

// handleSignup processes the signup request
func (h *AuthHandler) handleSignup(c echo.Context) error {
	signup := &user.Signup{
		Email:     c.FormValue("email"),
		Password:  c.FormValue("password"),
		FirstName: c.FormValue("first_name"),
		LastName:  c.FormValue("last_name"),
	}

	if _, err := h.UserService.SignUp(c.Request().Context(), signup); err != nil {
		h.Logger.Error("signup failed", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusBadRequest, "Failed to create user")
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}

// handleLogout processes the logout request
func (h *AuthHandler) handleLogout(c echo.Context) error {
	// Clear session cookie via SessionManager
	h.SessionManager.ClearSessionCookie(c)
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
