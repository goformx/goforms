package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// SettingsHandler handles settings-related requests
type SettingsHandler struct {
	HandlerDeps
}

// NewSettingsHandler creates a new settings handler using HandlerDeps
func NewSettingsHandler(deps HandlerDeps) (*SettingsHandler, error) {
	if err := deps.Validate("BaseHandler", "UserService", "SessionManager", "Renderer", "MiddlewareManager", "Config", "Logger"); err != nil {
		return nil, err
	}
	return &SettingsHandler{HandlerDeps: deps}, nil
}

// Register registers the settings routes
func (h *SettingsHandler) Register(e *echo.Echo) {
	e.GET("/settings", h.handleSettings)
}

// handleSettings handles the settings page request
func (h *SettingsHandler) handleSettings(c echo.Context) error {
	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(uint)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get user object
	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		h.Logger.Error("failed to get user (nil or error)", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user")
	}

	data := shared.BuildPageData(h.Config, "Settings")
	data.User = user
	data.Content = pages.SettingsContent(data)
	return h.Renderer.Render(c, pages.Settings(data))
}
