package handler

import (
	"errors"
	"net/http"
	"time"

	amw "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

const (
	// CookieMaxAgeMinutes is the number of minutes before a cookie expires
	CookieMaxAgeMinutes = 15
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	*handlers.BaseHandler
	renderer          *view.Renderer
	middlewareManager *amw.Manager
	config            *config.Config
	userService       user.Service
	sessionManager    *amw.SessionManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	baseHandler *handlers.BaseHandler,
	userService user.Service,
	sessionManager *amw.SessionManager,
	renderer *view.Renderer,
	middlewareManager *amw.Manager,
	config *config.Config,
) *AuthHandler {
	return &AuthHandler{
		BaseHandler:       baseHandler,
		userService:       userService,
		sessionManager:    sessionManager,
		renderer:          renderer,
		middlewareManager: middlewareManager,
		config:            config,
	}
}

// Register registers the auth routes
func (h *AuthHandler) Register(e *echo.Echo) {
	// Auth routes
	e.POST("/login", h.handleLoginPost)
	e.POST("/signup", h.handleSignupPost)
	e.POST("/logout", h.handleLogout)

	// Auth validation routes
	e.GET("/api/validation/login", h.handleLoginValidation)
	e.GET("/api/validation/signup", h.handleSignupValidation)
}

// handleLoginPost handles the login form submission
func (h *AuthHandler) handleLoginPost(c echo.Context) error {
	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Create login request
	login := &user.Login{
		Email:    email,
		Password: password,
	}

	// Attempt login
	loginResp, loginErr := h.userService.Login(c.Request().Context(), login)
	if loginErr != nil {
		// Handle specific error types
		switch {
		case errors.Is(loginErr, user.ErrInvalidCredentials):
			data := shared.PageData{
				Title:     "Login - GoFormX",
				CSRFToken: c.Get("csrf").(string),
				AssetPath: web.GetAssetPath,
			}
			return c.Render(
				http.StatusBadRequest,
				"login",
				pages.LoginWithError(data, "Invalid email or password"),
			)

		default:
			// Log unexpected errors
			h.LogError("failed to login", loginErr)
			data := shared.PageData{
				Title:     "Login - GoFormX",
				CSRFToken: c.Get("csrf").(string),
				AssetPath: web.GetAssetPath,
			}
			return c.Render(
				http.StatusInternalServerError,
				"login",
				pages.LoginWithError(data, "An error occurred. Please try again."),
			)
		}
	}

	// Create session
	sessionID, sessionErr := h.sessionManager.CreateSession(
		loginResp.User.ID, loginResp.User.Email, loginResp.User.Role,
	)
	if sessionErr != nil {
		h.LogError("failed to create session", sessionErr)
		data := shared.PageData{
			Title:     "Login - GoFormX",
			CSRFToken: c.Get("csrf").(string),
			AssetPath: web.GetAssetPath,
		}
		return c.Render(
			http.StatusInternalServerError,
			"login",
			pages.LoginWithError(data, "An error occurred. Please try again."),
		)
	}

	// Set session cookie
	h.sessionManager.SetSessionCookie(c, sessionID)

	// Set refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    loginResp.Token.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(CookieMaxAgeMinutes * time.Minute.Seconds()),
	})

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// handleSignupPost handles the signup form submission
func (h *AuthHandler) handleSignupPost(c echo.Context) error {
	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")
	firstName := c.FormValue("first_name")
	lastName := c.FormValue("last_name")

	// Validate password confirmation
	if password != confirmPassword {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Passwords do not match",
		})
	}

	// Create signup request
	signup := &user.Signup{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	// Attempt signup
	u, signupErr := h.userService.SignUp(c.Request().Context(), signup)
	if signupErr != nil {
		if errors.Is(signupErr, user.ErrUserExists) || errors.Is(signupErr, user.ErrEmailAlreadyExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Email already exists",
			})
		}
		h.LogError("failed to signup", signupErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "An error occurred. Please try again.",
		})
	}

	// Create session
	sessionID, sessionErr := h.sessionManager.CreateSession(u.ID, u.Email, u.Role)
	if sessionErr != nil {
		h.LogError("failed to create session", sessionErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "An error occurred. Please try again.",
		})
	}

	// Set session cookie
	h.sessionManager.SetSessionCookie(c, sessionID)

	// Return success response
	return c.JSON(http.StatusOK, map[string]string{
		"message":  "Signup successful",
		"redirect": "/dashboard",
	})
}

// handleLogout handles the logout request
func (h *AuthHandler) handleLogout(c echo.Context) error {
	// Get session ID from cookie
	cookie, err := c.Cookie("session_id")
	if err == nil {
		// Delete session
		h.sessionManager.DeleteSession(cookie.Value)
	}

	// Clear session cookie
	h.sessionManager.ClearSessionCookie(c)

	// Return success response
	return c.JSON(http.StatusOK, map[string]string{
		"message":  "Logout successful",
		"redirect": "/",
	})
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
			"min":     8,
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
			"min":     8,
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
