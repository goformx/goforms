package web

import (
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/context"
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

	// API routes
	api := e.Group("/api/v1")
	validation := api.Group("/validation")
	validation.GET("/login", h.LoginValidation)
	validation.GET("/signup", h.SignupValidation)
}

// generateValidationSchema generates a validation schema from struct tags
func generateValidationSchema(s interface{}) map[string]any {
	schema := make(map[string]any)
	t := reflect.TypeOf(s)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only process structs
	if t.Kind() != reflect.Struct {
		return schema
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Get the first part of the JSON tag (before any comma)
		fieldName := strings.Split(jsonTag, ",")[0]
		validateTag := field.Tag.Get("validate")

		// Create field schema
		fieldSchema := make(map[string]any)
		fieldSchema["type"] = "string" // Default type

		// Parse validation tags
		if validateTag != "" {
			rules := strings.Split(validateTag, ",")
			for _, rule := range rules {
				switch {
				case rule == "required":
					fieldSchema["min"] = 1
					fieldSchema["message"] = fieldName + " is required"
				case rule == "email":
					fieldSchema["type"] = "email"
					fieldSchema["message"] = "Please enter a valid email address"
				case strings.HasPrefix(rule, "min="):
					min := strings.TrimPrefix(rule, "min=")
					fieldSchema["min"] = min
					if fieldName == "password" {
						fieldSchema["message"] = "Password must be at least " + min + " characters long"
					}
				case strings.HasPrefix(rule, "match="):
					matchField := strings.TrimPrefix(rule, "match=")
					fieldSchema["type"] = "match"
					fieldSchema["matchField"] = matchField
					fieldSchema["message"] = "Passwords must match"
				}
			}
		}

		schema[fieldName] = fieldSchema
	}

	return schema
}

// Login handles GET /login - displays the login form
func (h *AuthHandler) Login(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, c, "Login")
	if context.IsAuthenticated(c) {
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}
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
		h.deps.Logger.Debug("Login failed",
			"email", h.deps.Logger.SanitizeField("email", email),
			"error_type", "authentication_error")
		data.Message = &shared.Message{
			Type: "error",
			Text: "Invalid email or password",
		}
		return h.deps.Renderer.Render(c, pages.Login(data))
	}

	// Create session with rolling expiration
	session, err := h.deps.SessionManager.CreateSession(
		authenticatedUser.ID,
		authenticatedUser.Email,
		c.Request().UserAgent(),
	)
	if err != nil {
		h.deps.Logger.Error("Failed to create session",
			"error", err,
			"user_id", h.deps.Logger.SanitizeField("user_id", authenticatedUser.ID))
		data.Message = &shared.Message{
			Type: "error",
			Text: "An error occurred. Please try again.",
		}
		return h.deps.Renderer.Render(c, pages.Login(data))
	}

	// Set session cookie using session manager
	h.deps.SessionManager.SetSessionCookie(c, session)

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Signup handles GET /signup - displays the signup form
func (h *AuthHandler) Signup(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, c, "Sign Up")
	if context.IsAuthenticated(c) {
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}
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

	// Create user
	signup := &user.Signup{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	newUser, err := h.deps.UserService.SignUp(c.Request().Context(), signup)
	if err != nil {
		h.deps.Logger.Debug("Signup failed",
			"error", err,
			"email", h.deps.Logger.SanitizeField("email", email))
		data.Message = &shared.Message{
			Type: "error",
			Text: "Unable to create account. Please try again.",
		}
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	// Create session for new user
	session, err := h.deps.SessionManager.CreateSession(newUser.ID, newUser.Email, c.Request().UserAgent())
	if err != nil {
		h.deps.Logger.Error("Failed to create session after signup",
			"error", err,
			"user_id", h.deps.Logger.SanitizeField("user_id", newUser.ID))
		data.Message = &shared.Message{
			Type: "error",
			Text: "An error occurred. Please try again.",
		}
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	// Set session cookie using session manager
	h.deps.SessionManager.SetSessionCookie(c, session)

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
	// Set content type to JSON
	c.Response().Header().Set("Content-Type", "application/json")
	schema := generateValidationSchema(&user.Login{})
	return c.JSON(http.StatusOK, schema)
}

// SignupValidation returns the validation schema for the signup form
func (h *AuthHandler) SignupValidation(c echo.Context) error {
	// Set content type to JSON
	c.Response().Header().Set("Content-Type", "application/json")
	schema := generateValidationSchema(&user.Signup{})
	return c.JSON(http.StatusOK, schema)
}
