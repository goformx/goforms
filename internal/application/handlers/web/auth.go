package web

import (
	"net/http"
	"time"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
	"github.com/mrz1836/go-sanitize"
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

	// Validation schema endpoints
	e.GET("/api/validation/login", h.handleLoginValidation)
	e.GET("/api/validation/signup", h.handleSignupValidation)
}

// Login handles the login page request
func (h *AuthHandler) Login(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, "Login")
	// Debug log for environment and asset path
	if h.deps.Config != nil && h.deps.Logger != nil {
		h.deps.Logger.Debug("Rendering login page",
			"env", h.deps.Config.App.Env,
			"assetPath", data.AssetPath("src/js/login.ts"),
		)
	}
	return h.deps.Renderer.Render(c, pages.Login(data))
}

// LoginPost handles the login form submission
func (h *AuthHandler) LoginPost(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Validate credentials using the user service
	user, err := h.deps.UserService.Authenticate(c.Request().Context(), email, password)
	if err != nil {
		h.deps.Logger.Debug("Login failed", "error", err)
		data := shared.BuildPageData(h.deps.Config, "Login")
		data.Error = "Invalid email or password"
		return h.deps.Renderer.Render(c, pages.Login(data))
	}

	// Create session
	session, err := h.deps.SessionManager.CreateSession(user.ID, user.Email, c.Request().UserAgent())
	if err != nil {
		h.deps.Logger.Error("Failed to create session", "error", err)
		data := shared.BuildPageData(h.deps.Config, "Login")
		data.Error = "An error occurred. Please try again."
		return h.deps.Renderer.Render(c, pages.Login(data))
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     h.deps.SessionManager.GetCookieName(),
		Value:    session,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours in seconds
	}
	c.SetCookie(cookie)

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Signup handles the signup page request
func (h *AuthHandler) Signup(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, "Sign Up")
	return h.deps.Renderer.Render(c, pages.Signup(data))
}

// SignupPost handles the signup form submission
func (h *AuthHandler) SignupPost(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	firstName := c.FormValue("first_name")
	lastName := c.FormValue("last_name")

	// Sanitize input (per validation rules)
	email = sanitize.Email(email, false)
	firstName = sanitize.XSS(firstName)
	lastName = sanitize.XSS(lastName)

	// Validate input (basic check)
	if email == "" || password == "" || firstName == "" || lastName == "" {
		data := shared.BuildPageData(h.deps.Config, "Sign Up")
		data.Error = "All fields are required."
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	signup := &user.Signup{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	newUser, err := h.deps.UserService.SignUp(c.Request().Context(), signup)
	if err != nil {
		h.deps.Logger.Debug("Signup failed", "error", err)
		data := shared.BuildPageData(h.deps.Config, "Sign Up")
		data.Error = "Signup failed: " + err.Error()
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	// Create session for new user
	session, err := h.deps.SessionManager.CreateSession(newUser.ID, newUser.Email, c.Request().UserAgent())
	if err != nil {
		h.deps.Logger.Error("Failed to create session after signup", "error", err)
		data := shared.BuildPageData(h.deps.Config, "Sign Up")
		data.Error = "An error occurred. Please try again."
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     h.deps.SessionManager.GetCookieName(),
		Value:    session,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours in seconds
	}
	c.SetCookie(cookie)

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
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
