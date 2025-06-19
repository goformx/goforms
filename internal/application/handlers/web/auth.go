package web

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/auth"
	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/request"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
	"github.com/mrz1836/go-sanitize"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	*BaseHandler
	AuthMiddleware *auth.Middleware
	RequestUtils   *request.Utils
	SchemaGenerator *validation.SchemaGenerator
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
) (*AuthHandler, error) {
	if base == nil {
		return nil, errors.New("base handler cannot be nil")
	}

	if authMiddleware == nil {
		return nil, errors.New("auth middleware cannot be nil")
	}

	if requestUtils == nil {
		return nil, errors.New("request utils cannot be nil")
	}

	if schemaGenerator == nil {
		return nil, errors.New("schema generator cannot be nil")
	}

	return &AuthHandler{
		BaseHandler:     base,
		AuthMiddleware:  authMiddleware,
		RequestUtils:    requestUtils,
		SchemaGenerator: schemaGenerator,
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
	h.Logger.Info("TestEndpoint called")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Test endpoint working",
		"status":  "success",
	})
}

// Login handles GET /login - displays the login form
func (h *AuthHandler) Login(c echo.Context) error {
	data := shared.BuildPageData(h.Config, c, "Login")
	if mwcontext.IsAuthenticated(c) {
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}
	// Debug log for environment and asset path
	if h.Config != nil && h.Logger != nil {
		h.Logger.Debug("Rendering login page",
			"env", h.Config.App.Env,
			"assetPath", data.AssetPath("src/js/login.ts"),
		)
	}
	return h.Renderer.Render(c, pages.Login(data))
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
			h.Logger.Error("failed to parse JSON request", "error", err)

			// Check if this is an AJAX request
			if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "Invalid request format",
				})
			}

			// For regular form submissions, render the login page with error
			data := shared.BuildPageData(h.Config, c, "Login")
			data.Message = &shared.Message{
				Type: "error",
				Text: "Invalid request format",
			}
			return h.Renderer.Render(c, pages.Login(data))
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
	loginResp, err := h.UserService.Login(c.Request().Context(), &user.Login{
		Email:    email,
		Password: password,
	})
	if err != nil {
		h.Logger.Error("login failed", "error", err)

		// Check if this is an AJAX request
		if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Invalid email or password",
			})
		}

		// For regular form submissions, render the login page with error
		data := shared.BuildPageData(h.Config, c, "Login")
		data.Message = &shared.Message{
			Type: "error",
			Text: "Invalid email or password",
		}
		return h.Renderer.Render(c, pages.Login(data))
	}

	// Create session
	session, err := h.SessionManager.CreateSession(loginResp.User.ID, loginResp.User.Email, c.Request().UserAgent())
	if err != nil {
		h.Logger.Error("failed to create session", "error", err)

		// Check if this is an AJAX request
		if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to create session",
			})
		}

		// For regular form submissions, render the login page with error
		data := shared.BuildPageData(h.Config, c, "Login")
		data.Message = &shared.Message{
			Type: "error",
			Text: "Failed to create session. Please try again.",
		}
		return h.Renderer.Render(c, pages.Login(data))
	}

	// Set session cookie
	h.SessionManager.SetSessionCookie(c, session)

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
	data := shared.BuildPageData(h.Config, c, "Sign Up")
	if mwcontext.IsAuthenticated(c) {
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}
	return h.Renderer.Render(c, pages.Signup(data))
}

// SignupPost handles the signup form submission
func (h *AuthHandler) SignupPost(c echo.Context) error {
	var signup user.Signup

	// Check content type to determine how to parse the request
	contentType := c.Request().Header.Get("Content-Type")
	if contentType == "application/json" {
		// Parse JSON request directly into signup struct
		if err := c.Bind(&signup); err != nil {
			h.Logger.Error("failed to parse JSON request", "error", err)

			// Check if this is an AJAX request
			if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "Invalid request format",
				})
			}

			// For regular form submissions, render the signup page with error
			data := shared.BuildPageData(h.Config, c, "Sign Up")
			data.Message = &shared.Message{
				Type: "error",
				Text: "Invalid request format",
			}
			return h.Renderer.Render(c, pages.Signup(data))
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
	newUser, err := h.UserService.SignUp(c.Request().Context(), &signup)
	if err != nil {
		h.Logger.Error("failed to create user", "error", err)

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
		data := shared.BuildPageData(h.Config, c, "Sign Up")
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
		return h.Renderer.Render(c, pages.Signup(data))
	}

	// Create session for new user
	session, err := h.SessionManager.CreateSession(newUser.ID, newUser.Email, c.Request().UserAgent())
	if err != nil {
		h.Logger.Error("failed to create session", "error", err)

		// Check if this is an AJAX request
		if c.Request().Header.Get("X-Requested-With") == XMLHttpRequestHeader {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Account created but unable to sign in. Please try logging in.",
			})
		}

		// For regular form submissions, render the signup page with error
		data := shared.BuildPageData(h.Config, c, "Sign Up")
		data.Message = &shared.Message{
			Type: "error",
			Text: "Account created but unable to sign in. Please try logging in.",
		}
		return h.Renderer.Render(c, pages.Signup(data))
	}

	// Set session cookie
	h.SessionManager.SetSessionCookie(c, session)

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
	cookie, err := c.Cookie(h.SessionManager.GetCookieName())
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Delete session
	h.SessionManager.DeleteSession(cookie.Value)

	// Clear session cookie
	h.SessionManager.ClearSessionCookie(c)

	return c.Redirect(http.StatusSeeOther, "/login")
}

// LoginValidation handles the login form validation schema request
func (h *AuthHandler) LoginValidation(c echo.Context) error {
	h.Logger.Info("LoginValidation endpoint called",
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"user_agent", c.Request().UserAgent(),
		"remote_addr", c.RealIP())

	// Generate schema using the validation package
	schema := h.SchemaGenerator.GenerateLoginSchema()

	h.Logger.Info("Generated login validation schema successfully", "schema", schema)
	return c.JSON(http.StatusOK, schema)
}

// SignupValidation returns the validation schema for the signup form
func (h *AuthHandler) SignupValidation(c echo.Context) error {
	h.Logger.Info("SignupValidation endpoint called",
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"user_agent", c.Request().UserAgent(),
		"remote_addr", c.RealIP())

	// Generate schema using the validation package
	schema := h.SchemaGenerator.GenerateSignupSchema()

	h.Logger.Info("Generated signup validation schema successfully", "schema", schema)
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
