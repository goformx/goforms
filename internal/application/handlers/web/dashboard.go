package web

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
)

type DashboardHandler struct {
	*BaseHandler
	AccessManager *access.AccessManager
}

func NewDashboardHandler(base *BaseHandler, accessManager *access.AccessManager) *DashboardHandler {
	return &DashboardHandler{
		BaseHandler:   base,
		AccessManager: accessManager,
	}
}

func (h *DashboardHandler) Register(e *echo.Echo) {
	// Create dashboard group with access control
	dashboard := e.Group(constants.PathDashboard)
	dashboard.Use(access.Middleware(h.AccessManager, h.Logger))
	dashboard.GET("", h.handleDashboard)
}

// handleDashboard handles the dashboard page request
func (h *DashboardHandler) handleDashboard(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return h.HandleForbidden(c, "User not authenticated")
	}

	// Get user data
	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get user data", "error", err)
		return h.HandleError(c, err, "Failed to get user data")
	}

	// Get forms for the user
	forms, err := h.FormService.ListForms(c.Request().Context(), userID)
	if err != nil {
		h.Logger.Error("failed to list forms", "error", err)
		return h.HandleError(c, err, "Failed to list forms")
	}

	// Build page data
	data := view.BuildPageData(h.Config, h.AssetManager, c, "Dashboard")
	data.User = user
	data.Forms = forms

	// Render dashboard template
	return h.Renderer.Render(c, pages.Dashboard(data, forms))
}

// Start initializes the dashboard handler.
// This is called during application startup.
func (h *DashboardHandler) Start(ctx context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the dashboard handler.
// This is called during application shutdown.
func (h *DashboardHandler) Stop(ctx context.Context) error {
	return nil // No cleanup needed
}
