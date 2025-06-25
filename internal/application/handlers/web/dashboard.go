package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
)

// DashboardHandler handles dashboard routes.
type DashboardHandler struct {
	*BaseHandler
	AccessManager  *access.Manager
	AuthMiddleware *auth.Middleware
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(
	base *BaseHandler,
	accessManager *access.Manager,
	authMiddleware *auth.Middleware,
) *DashboardHandler {
	return &DashboardHandler{
		BaseHandler:    base,
		AccessManager:  accessManager,
		AuthMiddleware: authMiddleware,
	}
}

// handleDashboard handles the dashboard page request
func (h *DashboardHandler) handleDashboard(c echo.Context) error {
	// Get user from context using the auth middleware helper
	user, ok := h.AuthMiddleware.GetUserFromContext(c)
	if !ok {
		// This should not happen if auth middleware is working correctly
		h.Logger.Error("user not found in context despite authentication")
		return c.Redirect(http.StatusSeeOther, constants.PathLogin)
	}

	// Get forms for the user
	forms, err := h.FormService.ListForms(c.Request().Context(), user.ID)
	if err != nil {
		h.Logger.Error("failed to list forms", "error", err)
		return h.HandleError(c, err, "Failed to list forms")
	}

	// Build page data
	data := view.BuildPageData(h.Config, h.AssetManager, c, "Dashboard")
	data.User = user
	data.Forms = forms

	// Render dashboard template
	if renderErr := h.Renderer.Render(c, pages.Dashboard(data, forms)); renderErr != nil {
		return fmt.Errorf("render dashboard: %w", renderErr)
	}
	return nil
}

// Start initializes the dashboard handler.
// This is called during application startup.
func (h *DashboardHandler) Start(_ context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the dashboard handler.
// This is called during application shutdown.
func (h *DashboardHandler) Stop(_ context.Context) error {
	return nil // No cleanup needed
}
