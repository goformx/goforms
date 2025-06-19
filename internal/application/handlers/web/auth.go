package web

import (
	"context"
	"errors"
	"net/http"
	"reflect"
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
// Note: Routes are actually registered by RegisterHandlers in module.go
func (h *AuthHandler) Register(e *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface

	// Add a simple test endpoint to verify JSON responses work
	api := e.Group("/api/v1")
	api.GET("/test", h.TestEndpoint)
}

// TestEndpoint is a simple test endpoint to verify JSON responses work
func (h *AuthHandler) TestEndpoint(c echo.Context) error {
	h.deps.Logger.Info("TestEndpoint called")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Test endpoint working",
		"status":  "success",
	})
}

func getFieldSchema(field reflect.StructField) map[string]any {
	fieldSchema := make(map[string]any)

	// Get validation tags
	validate := field.Tag.Get("validate")
	if validate != "" {
		fieldSchema["validate"] = validate
	}

	// Get min/max length
	minLen := field.Tag.Get("minlen")
	if minLen != "" {
		fieldSchema["minLength"] = minLen
	}
	maxLen := field.Tag.Get("maxlen")
	if maxLen != "" {
		fieldSchema["maxLength"] = maxLen
	}

	// Set type and message based on validation rules
	if validate != "" {
		if validate == "required,email" {
			fieldSchema["type"] = "email"
			fieldSchema["message"] = "Please enter a valid email address"
		} else if validate == "required" {
			fieldSchema["type"] = "string"
			fieldSchema["message"] = "This field is required"
		} else if validate == "required,min=8" {
			fieldSchema["type"] = "password"
			fieldSchema["min"] = "8"
			fieldSchema["message"] = "Password must be at least 8 characters long and include uppercase, lowercase, number, and special characters"
		} else if validate == "required,eqfield=password" {
			fieldSchema["type"] = "match"
			fieldSchema["matchField"] = "password"
			fieldSchema["message"] = "Passwords don't match"
		}
	}

	return fieldSchema
}

func generateValidationSchema(s any) map[string]any {
	t := reflect.TypeOf(s)
	schema := make(map[string]any)

	for i := range t.NumField() {
		field := t.Field(i)
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = field.Name
		}

		fieldSchema := getFieldSchema(field)
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

			// Check if this is an AJAX request
			if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "Invalid request format",
				})
			}

			// For regular form submissions, render the login page with error
			data := shared.BuildPageData(h.deps.Config, c, "Login")
			data.Message = &shared.Message{
				Type: "error",
				Text: "Invalid request format",
			}
			return h.deps.Renderer.Render(c, pages.Login(data))
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

		// Check if this is an AJAX request
		if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Invalid email or password",
			})
		}

		// For regular form submissions, render the login page with error
		data := shared.BuildPageData(h.deps.Config, c, "Login")
		data.Message = &shared.Message{
			Type: "error",
			Text: "Invalid email or password",
		}
		return h.deps.Renderer.Render(c, pages.Login(data))
	}

	// Create session
	session, err := h.deps.SessionManager.CreateSession(loginResp.User.ID, loginResp.User.Email, c.Request().UserAgent())
	if err != nil {
		h.deps.Logger.Error("failed to create session", "error", err)

		// Check if this is an AJAX request
		if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to create session",
			})
		}

		// For regular form submissions, render the login page with error
		data := shared.BuildPageData(h.deps.Config, c, "Login")
		data.Message = &shared.Message{
			Type: "error",
			Text: "Failed to create session. Please try again.",
		}
		return h.deps.Renderer.Render(c, pages.Login(data))
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

			// Check if this is an AJAX request
			if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "Invalid request format",
				})
			}

			// For regular form submissions, render the signup page with error
			data := shared.BuildPageData(h.deps.Config, c, "Sign Up")
			data.Message = &shared.Message{
				Type: "error",
				Text: "Invalid request format",
			}
			return h.deps.Renderer.Render(c, pages.Signup(data))
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

		// Check if this is an AJAX request
		if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
			// Check for specific error types
			if errors.Is(err, user.ErrUserExists) {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "This email is already registered. Please try signing in instead.",
					"field":   "email",
				})
			}

			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Unable to create account. Please try again.",
			})
		}

		// For regular form submissions, render the signup page with error
		data := shared.BuildPageData(h.deps.Config, c, "Sign Up")
		if errors.Is(err, user.ErrUserExists) {
			data.Message = &shared.Message{
				Type: "error",
				Text: "This email is already registered. Please try signing in instead.",
			}
		} else {
			data.Message = &shared.Message{
				Type: "error",
				Text: "Unable to create account. Please try again.",
			}
		}
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	// Create session for new user
	session, err := h.deps.SessionManager.CreateSession(newUser.ID, newUser.Email, c.Request().UserAgent())
	if err != nil {
		h.deps.Logger.Error("failed to create session", "error", err)

		// Check if this is an AJAX request
		if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Account created but unable to sign in. Please try logging in.",
			})
		}

		// For regular form submissions, render the signup page with error
		data := shared.BuildPageData(h.deps.Config, c, "Sign Up")
		data.Message = &shared.Message{
			Type: "error",
			Text: "Account created but unable to sign in. Please try logging in.",
		}
		return h.deps.Renderer.Render(c, pages.Signup(data))
	}

	// Set session cookie
	h.deps.SessionManager.SetSessionCookie(c, session)

	// Return success
	if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
		return c.JSON(http.StatusOK, map[string]string{
			"message":  "Account created successfully!",
			"redirect": "/dashboard",
		})
	}

	// Redirect to dashboard for regular requests
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
	h.deps.Logger.Info("LoginValidation endpoint called",
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"remote_addr", c.Request().RemoteAddr)

	// Set content type to JSON
	c.Response().Header().Set("Content-Type", "application/json")

	// Generate schema with simple error handling
	schema := generateValidationSchema(user.Login{})

	h.deps.Logger.Info("Generated login validation schema successfully", "schema", schema)
	return c.JSON(http.StatusOK, schema)
}

// SignupValidation returns the validation schema for the signup form
func (h *AuthHandler) SignupValidation(c echo.Context) error {
	h.deps.Logger.Info("SignupValidation endpoint called",
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"remote_addr", c.Request().RemoteAddr)

	// Set content type to JSON
	c.Response().Header().Set("Content-Type", "application/json")

	// Generate schema with simple error handling
	schema := generateValidationSchema(user.Signup{})

	h.deps.Logger.Info("Generated signup validation schema successfully", "schema", schema)
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
