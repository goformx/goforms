package web

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"time"

	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
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
	// XMLHttpRequestHeader is the standard header value for AJAX requests
	XMLHttpRequestHeader = "XMLHttpRequest"
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
func generateValidationSchema(s any) map[string]any {
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
					fieldSchema["type"] = "required"
					fieldSchema["message"] = fieldName + " is required"
				case rule == "email":
					fieldSchema["type"] = "email"
					fieldSchema["message"] = "Please enter a valid email address"
				case strings.HasPrefix(rule, "min="):
					minLength := strings.TrimPrefix(rule, "min=")
					fieldSchema["min"] = minLength
					if fieldName == "password" {
						fieldSchema["type"] = "password"
						fieldSchema["message"] = "Password must be at least " + minLength + " characters long"
					}
				case strings.HasPrefix(rule, "eqfield="):
					matchField := strings.TrimPrefix(rule, "eqfield=")
					fieldSchema["type"] = "match"
					fieldSchema["matchField"] = strings.ToLower(matchField)
					fieldSchema["message"] = "Passwords must match"
				}
			}
		}

		// Special handling for password fields
		if fieldName == "password" {
			fieldSchema["type"] = "password"
			if _, hasMin := fieldSchema["min"]; !hasMin {
				fieldSchema["min"] = MinPasswordLength
			}
			fieldSchema["message"] = "Password must be at least 8 characters long and include uppercase, lowercase, number, and special characters"
		}

		// Special handling for confirm_password
		if fieldName == "confirm_password" {
			fieldSchema["type"] = "match"
			fieldSchema["matchField"] = "password"
			fieldSchema["message"] = "Passwords must match"
			fieldSchema["min"] = MinPasswordLength
		}

		schema[fieldName] = fieldSchema
	}

	return schema
}

// Login handles GET /login - displays the login form
func (h *AuthHandler) Login(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, c, "Login")
	if mwcontext.IsAuthenticated(c) {
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

/**
 * LoginPost handles POST /login - processes the login form
 *
 * This handler:
 * 1. Validates user credentials
 * 2. Creates a new session on success
 * 3. Sets session cookie
 * 4. Returns appropriate response based on request type:
 *    - JSON response for API requests
 *    - HTML response with error for regular requests
 *    - Redirect to dashboard on success
 */
func (h *AuthHandler) LoginPost(c echo.Context) error {
	var (
		email    string
		password string
	)

	// Check content type to determine how to parse the request
	contentType := c.Request().Header.Get("Content-Type")
	if contentType == "application/json" {
		// Parse JSON request
		var data struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.Bind(&data); err != nil {
			h.deps.Logger.Error("failed to parse JSON request", "error", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid request format",
			})
		}
		email = data.Email
		password = data.Password
	} else {
		// Parse form data
		email = c.FormValue("email")
		password = c.FormValue("password")
	}

	// Sanitize email
	email = sanitize.Email(email, false)

	// Validate credentials
	loginResp, err := h.deps.UserService.Login(c.Request().Context(), &user.Login{
		Email:    email,
		Password: password,
	})
	if err != nil {
		h.deps.Logger.Error("login failed", "error", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "Invalid email or password",
		})
	}

	// Create session
	session, err := h.deps.SessionManager.CreateSession(loginResp.User.ID, loginResp.User.Email, c.Request().UserAgent())
	if err != nil {
		h.deps.Logger.Error("failed to create session", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to create session",
		})
	}

	// Set session cookie
	h.deps.SessionManager.SetSessionCookie(c, session)

	// Return success
	if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
		return c.JSON(http.StatusOK, map[string]string{
			"redirect": "/dashboard",
		})
	}

	// Redirect to dashboard for regular requests
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Signup handles GET /signup - displays the signup form
func (h *AuthHandler) Signup(c echo.Context) error {
	data := shared.BuildPageData(h.deps.Config, c, "Sign Up")
	if mwcontext.IsAuthenticated(c) {
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}
	return h.deps.Renderer.Render(c, pages.Signup(data))
}

// SignupPost handles the signup form submission
func (h *AuthHandler) SignupPost(c echo.Context) error {
	var signup user.Signup

	// Check content type to determine how to parse the request
	contentType := c.Request().Header.Get("Content-Type")
	if contentType == "application/json" {
		// Parse JSON request directly into signup struct
		if err := c.Bind(&signup); err != nil {
			h.deps.Logger.Error("failed to parse JSON request", "error", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid request format",
			})
		}
	} else {
		// Parse form data
		signup = user.Signup{
			Email:           c.FormValue("email"),
			Password:        c.FormValue("password"),
			ConfirmPassword: c.FormValue("confirm_password"),
		}
	}

	// Sanitize email
	signup.Email = sanitize.Email(signup.Email, false)

	// Create user
	newUser, err := h.deps.UserService.SignUp(c.Request().Context(), &signup)
	if err != nil {
		h.deps.Logger.Error("failed to create user", "error", err)

		// Check for specific error types
		if err == user.ErrUserExists {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "This email is already registered. Please try signing in instead.",
				"field":   "email",
			})
		}

		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Unable to create account. Please try again.",
		})
	}

	// Create session for new user
	session, err := h.deps.SessionManager.CreateSession(newUser.ID, newUser.Email, c.Request().UserAgent())
	if err != nil {
		h.deps.Logger.Error("failed to create session", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Account created but unable to sign in. Please try logging in.",
		})
	}

	// Set session cookie
	h.deps.SessionManager.SetSessionCookie(c, session)

	// Return success
	return c.JSON(http.StatusOK, map[string]string{
		"message":  "Account created successfully!",
		"redirect": "/dashboard",
	})
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

// Start initializes the auth handler.
// This is called during application startup.
func (h *AuthHandler) Start(ctx context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the auth handler.
// This is called during application shutdown.
func (h *AuthHandler) Stop(ctx context.Context) error {
	return nil // No cleanup needed
}
