package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	baseHandler       *BaseHandler
	userService       domain.UserService
	sessionManager    *middleware.SessionManager
	renderer          *view.Renderer
	middlewareManager *middleware.Manager
	config            *config.Config
	logger            logging.Logger
}

const (
	// SessionDuration is the duration for which a session remains valid
	SessionDuration = 24 * time.Hour
)

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	baseHandler *BaseHandler,
	userService domain.UserService,
	sessionManager *middleware.SessionManager,
	renderer *view.Renderer,
	middlewareManager *middleware.Manager,
	cfg *config.Config,
	logger logging.Logger,
) *AuthHandler {
	return &AuthHandler{
		baseHandler:       baseHandler,
		userService:       userService,
		sessionManager:    sessionManager,
		renderer:          renderer,
		middlewareManager: middlewareManager,
		config:            cfg,
		logger:            logger,
	}
}

// Validate validates the handler's dependencies
func (h *AuthHandler) Validate() error {
	if h.baseHandler == nil {
		return errors.New("base handler is required")
	}
	if h.userService == nil {
		return errors.New("user service is required")
	}
	if h.sessionManager == nil {
		return errors.New("session manager is required")
	}
	if h.renderer == nil {
		return errors.New("renderer is required")
	}
	if h.middlewareManager == nil {
		return errors.New("middleware manager is required")
	}
	if h.config == nil {
		return errors.New("config is required")
	}
	if h.logger == nil {
		return errors.New("logger is required")
	}
	return nil
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
	return h.renderer.Render(c, pages.Login(data))
}

// handleLogin processes the login request
func (h *AuthHandler) handleLogin(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Authenticate user
	userData, err := h.userService.Authenticate(c.Request().Context(), email, password)
	if err != nil {
		h.logger.Error("failed to authenticate user", logging.ErrorField("error", err))
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
	return h.renderer.Render(c, pages.Signup(data))
}

// handleSignup processes the signup request
func (h *AuthHandler) handleSignup(c echo.Context) error {
	signup := &user.Signup{
		Email:     c.FormValue("email"),
		Password:  c.FormValue("password"),
		FirstName: c.FormValue("first_name"),
		LastName:  c.FormValue("last_name"),
	}

	if _, err := h.userService.SignUp(c.Request().Context(), signup); err != nil {
		h.logger.Error("signup failed", logging.ErrorField("error", err))
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
