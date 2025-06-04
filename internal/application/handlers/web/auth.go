package web

import (
	"net/http"
	"strconv"
	"time"

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
)

// NewAuthHandler creates a new auth handler using HandlerDeps
func NewAuthHandler(deps HandlerDeps) (*AuthHandler, error) {
	if err := deps.Validate("BaseHandler", "UserService", "SessionManager", "Renderer", "MiddlewareManager", "Config", "Logger"); err != nil {
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
}

// showLoginPage renders the login page
func (h *AuthHandler) showLoginPage(c echo.Context) error {
	data := shared.PageData{
		Title: "Login",
	}
	return h.Renderer.Render(c, pages.Login(data))
}

// handleLogin processes the login request
func (h *AuthHandler) handleLogin(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Authenticate user
	userData, err := h.UserService.Authenticate(c.Request().Context(), email, password)
	if err != nil {
		h.Logger.Error("failed to authenticate user", logging.ErrorField("error", err))
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials",
		})
	}

	// Set session cookie
	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = strconv.FormatUint(uint64(userData.ID), 10)
	cookie.Expires = time.Now().Add(SessionDuration)
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(cookie)

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// showSignupPage renders the signup page
func (h *AuthHandler) showSignupPage(c echo.Context) error {
	data := shared.PageData{
		Title: "Sign Up",
	}
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
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to create user",
		})
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}

// handleLogout processes the logout request
func (h *AuthHandler) handleLogout(c echo.Context) error {
	// Clear session cookie
	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(cookie)

	return c.Redirect(http.StatusSeeOther, "/login")
}
