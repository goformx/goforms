package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	HandlerDeps
}

func NewDashboardHandler(deps HandlerDeps) *DashboardHandler {
	return &DashboardHandler{HandlerDeps: deps}
}

func (h *DashboardHandler) Register(e *echo.Echo) {
	// Create dashboard group with RequireAuth middleware
	dashboard := e.Group("/dashboard", middleware.RequireAuth(h.Logger))
	dashboard.GET("", h.handleDashboard)
	dashboard.GET("/forms/:id", h.handleFormView)
}

func (h *DashboardHandler) handleDashboard(c echo.Context) error {
	ctx := c.Request().Context()
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok || userIDRaw == "" {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userObj, err := h.UserService.GetUserByID(ctx, userIDRaw)
	if err != nil || userObj == nil {
		h.Logger.Error("failed to get user for dashboard", "error", err)
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	forms, err := h.FormService.GetUserForms(ctx, userIDRaw)
	if err != nil {
		h.Logger.Error("failed to get user forms for dashboard", "error", err)
		return c.String(http.StatusInternalServerError, "Failed to load forms")
	}
	data := shared.BuildPageData(nil, c, "Dashboard")
	data.User = userObj
	return h.Renderer.Render(c, pages.Dashboard(data, forms))
}

// handleFormView handles the form view page request
func (h *DashboardHandler) handleFormView(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}

	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get user object
	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		h.Logger.Error("failed to get user (nil or error)", "error", err)
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user")
	}

	// Get form
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", "error", err)
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if form.UserID != userID {
		return response.ErrorResponse(c, http.StatusForbidden, "You don't have permission to view this form")
	}

	data := shared.BuildPageData(h.Config, c, "View Form")
	data.User = user
	data.Form = form
	return h.Renderer.Render(c, pages.Forms(data))
}
