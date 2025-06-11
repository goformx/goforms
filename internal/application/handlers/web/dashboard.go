package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/common/errors"
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
		h.Logger.Error("invalid user session",
			"error", "missing or invalid user_id",
			"session_id", c.Request().Header.Get("X-Session-ID"))
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	userObj, err := h.UserService.GetUserByID(ctx, userIDRaw)
	if err != nil {
		h.Logger.Error("failed to get user for dashboard",
			"error", err,
			"user_id", userIDRaw,
			"error_type", "user_service_error")

		if errors.IsNotFound(err) {
			return response.WebErrorResponse(c, h.Renderer, http.StatusNotFound, "User not found")
		}
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to load user data")
	}

	forms, err := h.FormService.GetUserForms(ctx, userIDRaw)
	if err != nil {
		h.Logger.Error("failed to get user forms for dashboard",
			"error", err,
			"user_id", userIDRaw,
			"error_type", "form_service_error")
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to load forms")
	}

	data := shared.BuildPageData(h.Config, c, "Dashboard")
	data.User = userObj
	return h.Renderer.Render(c, pages.Dashboard(data, forms))
}

// handleFormView handles the form view page request
func (h *DashboardHandler) handleFormView(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		h.Logger.Error("invalid form request",
			"error", "missing form id",
			"path", c.Request().URL.Path)
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok {
		h.Logger.Error("invalid user session",
			"error", "missing user_id",
			"session_id", c.Request().Header.Get("X-Session-ID"))
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get user object
	user, err := h.UserService.GetUserByID(c.Request().Context(), userIDRaw)
	if err != nil {
		h.Logger.Error("failed to get user for form view",
			"error", err,
			"user_id", userIDRaw,
			"form_id", formID,
			"error_type", "user_service_error")

		if errors.IsNotFound(err) {
			return response.WebErrorResponse(c, h.Renderer, http.StatusNotFound, "User not found")
		}
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}

	// Get form
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form",
			"error", err,
			"user_id", userIDRaw,
			"form_id", formID,
			"error_type", "form_service_error")

		if errors.IsNotFound(err) {
			return response.WebErrorResponse(c, h.Renderer, http.StatusNotFound, "Form not found")
		}
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if form.UserID != userIDRaw {
		h.Logger.Error("unauthorized form access attempt",
			"user_id", userIDRaw,
			"form_id", formID,
			"form_owner", form.UserID,
			"error_type", "authorization_error")
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to view this form")
	}

	data := shared.BuildPageData(h.Config, c, "View Form")
	data.User = user
	data.Form = form
	return h.Renderer.Render(c, pages.Forms(data))
}
