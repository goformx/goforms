package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
)

const (
	// CookieExpiryMinutes is the number of minutes before a cookie expires
	CookieExpiryMinutes = 15
	// SecondsInMinute is the number of seconds in a minute
	SecondsInMinute = 60
)

// AuthHandlerOption defines an auth handler option
type AuthHandlerOption func(*AuthHandler)

// WithUserService sets the user service
func WithUserService(svc user.Service) AuthHandlerOption {
	return func(h *AuthHandler) {
		h.userService = svc
	}
}

// AuthHandler handles authentication related requests
type AuthHandler struct {
	*handlers.BaseHandler
	userService user.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger logging.Logger, opts ...AuthHandlerOption) *AuthHandler {
	h := &AuthHandler{
		BaseHandler: handlers.NewBaseHandler(nil, nil, logger),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Validate validates that required dependencies are set
func (h *AuthHandler) Validate() error {
	if err := h.BaseHandler.Validate(); err != nil {
		h.LogError("failed to validate handler", err)
		return err
	}
	if h.userService == nil {
		return errors.New("user service is required")
	}
	return nil
}

// Register registers the auth routes
func (h *AuthHandler) Register(e *echo.Echo) {
	if err := h.Validate(); err != nil {
		h.LogError("failed to validate handler", err)
		return
	}

	// Web routes - logout only via POST for security
	e.POST("/logout", h.handleWebLogout)
}

// handleWebLogout handles web logout
func (h *AuthHandler) handleWebLogout(c echo.Context) error {
	// Get refresh token from cookie
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		h.LogError("failed to get refresh token cookie", err)
		return echo.NewHTTPError(http.StatusBadRequest, "No refresh token found")
	}

	// Blacklist the refresh token
	logoutErr := h.userService.Logout(c.Request().Context(), cookie.Value)
	if logoutErr != nil {
		h.LogError("failed to logout", logoutErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout")
	}

	// Clear the refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	return c.Redirect(http.StatusSeeOther, "/login")
}
