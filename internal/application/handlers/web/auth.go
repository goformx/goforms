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
	e.GET("/api/validation/login", h.LoginValidation)
	e.GET("/api/validation/signup", h.SignupValidation)
}

// Login handles GET /login - displays the login form
func (h *AuthHandler) Login(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, c, "Login")
	// Debug log for environment and asset path
	if h.deps.Config != nil && h.deps.Logger != nil {
		h.deps.Logger.Debug("Rendering login page",
			"env", h.deps.Config.App.Env,
			"assetPath", data.AssetPath("src/js/login.ts"),
		)
	}
	return h.deps.Renderer.Render(c, pages.Login(data))
}

// LoginPost handles POST /login - processes the login form
func (h *AuthHandler) LoginPost(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, c, "Login")

	email := c.FormValue("email")
	password := c.FormValue("password")

	// Validate credentials using the user service
	authenticatedUser, err := h.deps.UserService.Authenticate(c.Request().Context(), email, password)
	if err != nil {
		h.deps.Logger.Debug("Login failed", "error", err)
		data.Message = &shared.Message{
			Type: "error",
			Text: err.Error(),
		}
		return h.deps.Renderer.Render(c, pages.Login(data))
	}

	// Create session
	session, err := h.deps.SessionManager.CreateSession(
		authenticatedUser.ID,
		authenticatedUser.Email,
		c.Request().UserAgent(),
	)
	if err != nil {
		h.deps.Logger.Error("Failed to create session", "error", err)
		data.Message = &shared.Message{
			Type: "error",
			Text: "An error occurred. Please try again.",
		}
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
		MaxAge:   int(SessionDuration.Seconds()),
	}
	c.SetCookie(cookie)

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Signup handles GET /signup - displays the signup form
func (h *AuthHandler) Signup(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, c, "Sign Up")
	return h.deps.Renderer.Render(c, pages.Signup(data))
}

// SignupPost handles POST /signup - processes the signup form
func (h *AuthHandler) SignupPost(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, c, "Sign Up")

	// Get and sanitize form values
	email := sanitize.Email(c.FormValue("email"), false)
	password := c.FormValue("password")
	firstName := sanitize.XSS(c.FormValue("first_name"))
	lastName := sanitize.XSS(c.FormValue("last_name"))

	// Validate input
	if email == "" || password == "" || firstName == "" || lastName == "" {
		data.Message = &shared.Message{
			Type: "error",
			Text: "All fields are required.",
		}
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	// Create user
	signup := &user.Signup{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	newUser, err := h.deps.UserService.SignUp(c.Request().Context(), signup)
	if err != nil {
		h.deps.Logger.Debug("Signup failed", "error", err)
		data.Message = &shared.Message{
			Type: "error",
			Text: "Signup failed: " + err.Error(),
		}
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	// Create session for new user
	session, err := h.deps.SessionManager.CreateSession(newUser.ID, newUser.Email, c.Request().UserAgent())
	if err != nil {
		h.deps.Logger.Error("Failed to create session after signup", "error", err)
		data.Message = &shared.Message{
			Type: "error",
			Text: "An error occurred. Please try again.",
		}
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
		MaxAge:   int(SessionDuration.Seconds()),
	}
	c.SetCookie(cookie)

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Logout handles POST /logout - processes the logout request
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

// LoginValidation handles the login form validation schema request
func (h *AuthHandler) LoginValidation(c echo.Context) error {
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

// SignupValidation returns the validation schema for the signup form
func (h *AuthHandler) SignupValidation(c echo.Context) error {
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
