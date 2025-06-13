package web

import (
	"context"
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware/access"
	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	HandlerDeps
	AccessManager *access.AccessManager
}

func NewDashboardHandler(deps HandlerDeps, accessManager *access.AccessManager) *DashboardHandler {
	return &DashboardHandler{
		HandlerDeps:   deps,
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
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Fetch user data
	userObj, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || userObj == nil {
		// Sanitize and limit path length
		path := c.Request().URL.Path
		if len(path) > 100 {
			path = path[:100] + "..."
		}
		// Only log essential information
		h.Logger.Error("authentication error",
			"error", "user not found",
			"path", path,
		)
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Fetch forms data
	forms, err := h.FormService.ListForms(c.Request().Context(), map[string]any{"user_id": userID})
	if err != nil {
		h.Logger.Error("failed to get user forms", "user_id", userID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load forms"})
	}

	// Build page data
	data := shared.BuildPageData(h.Config, c, "Dashboard")
	data.User = userObj
	data.Forms = forms

	return h.Renderer.Render(c, pages.Dashboard(data, forms))
}

// handleFormView handles the form view page request
func (h *DashboardHandler) handleFormView(c echo.Context) error {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Fetch user data
	userObj, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || userObj == nil {
		h.Logger.Error("user not found after authentication", "user_id", userID, "path", c.Request().URL.Path)
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Fetch form data
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil || form == nil {
		h.Logger.Error("form not found or error loading form", "form_id", formID, "user_id", userID, "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusNotFound, "Resource not found")
	}

	// Verify form ownership
	if form.UserID != userID {
		h.Logger.Error("unauthorized form access attempt",
			"user_id", h.Logger.SanitizeField("user_id", userID),
			"form_id", h.Logger.SanitizeField("form_id", formID),
			"form_owner", h.Logger.SanitizeField("form_owner", form.UserID),
			"error_type", "authorization_error")
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to view this form")
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
