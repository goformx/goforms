package web

import (
	"context"

	"github.com/goformx/goforms/internal/application/middleware/access"
	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
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
	dashboard := e.Group("/dashboard")
	dashboard.Use(access.Middleware(h.AccessManager, h.Logger))
	dashboard.GET("", h.handleDashboard)
	dashboard.GET("/forms/:id", h.handleFormView)
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
	data := shared.BuildPageData(h.Config, c, "Dashboard")
	data.User = user
	data.Forms = forms

	// Render dashboard template
	return h.Renderer.Render(c, pages.Dashboard(data, forms))
}

// handleFormView handles the form view page request
func (h *DashboardHandler) handleFormView(c echo.Context) error {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return h.HandleForbidden(c, "User not authenticated")
	}

	formID := c.Param("id")
	if formID == "" {
		return h.HandleError(c, nil, "Form ID is required")
	}

	// Fetch user data
	userObj, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || userObj == nil {
		h.Logger.Error("user not found after authentication",
			"user_id", h.Logger.SanitizeField("user_id", userID),
			"path", h.Logger.SanitizeField("path", c.Request().URL.Path))
		return h.HandleNotFound(c, "User not found")
	}

	// Fetch form data
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil || form == nil {
		h.Logger.Error("form not found or error loading form",
			"form_id", h.Logger.SanitizeField("form_id", formID),
			"user_id", h.Logger.SanitizeField("user_id", userID),
			"error", err)
		return h.HandleNotFound(c, "Resource not found")
	}

	// Verify form ownership
	if form.UserID != userID {
		h.Logger.Error("unauthorized form access attempt",
			"user_id", h.Logger.SanitizeField("user_id", userID),
			"form_id", h.Logger.SanitizeField("form_id", formID),
			"form_owner", h.Logger.SanitizeField("form_owner", form.UserID),
			"error_type", "authorization_error")
		return h.HandleForbidden(c, "You don't have permission to view this form")
	}

	data := shared.BuildPageData(h.Config, c, "Form View")
	data.User = userObj
	data.Form = form
	return h.Renderer.Render(c, pages.Forms(data))
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
