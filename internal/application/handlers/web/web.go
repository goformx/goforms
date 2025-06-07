package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// WebHandler handles web page requests
type WebHandler struct {
	HandlerDeps
}

// NewWebHandler creates a new web handler using HandlerDeps
func NewWebHandler(deps HandlerDeps) (*WebHandler, error) {
	if err := deps.Validate(
		"BaseHandler",
		"UserService",
		"SessionManager",
		"Renderer",
		"MiddlewareManager",
		"Config",
		"Logger",
	); err != nil {
		return nil, err
	}
	return &WebHandler{HandlerDeps: deps}, nil
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	e.GET("/", h.handleHome)
	e.GET("/dashboard", h.handleDashboard)
	e.GET("/forms/:id", h.handleFormView)
}

// handleHome handles the home page request
func (h *WebHandler) handleHome(c echo.Context) error {
	data := shared.BuildPageData(h.Config, "Welcome to GoFormX")
	return h.Renderer.Render(c, pages.Home(data))
}

// handleDashboard handles the dashboard page request
func (h *WebHandler) handleDashboard(c echo.Context) error {
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

	// Get user's forms
	forms, err := h.BaseHandler.formService.GetUserForms(c.Request().Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get user forms", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get forms")
	}

	data := shared.BuildPageData(h.Config, "Dashboard")
	data.User = user
	data.Forms = forms
	return h.Renderer.Render(c, pages.Dashboard(data))
}

// handleFormView handles the form view page request
func (h *WebHandler) handleFormView(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}

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

	// Get form
	form, err := h.BaseHandler.formService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if form.UserID != userID {
		return response.ErrorResponse(c, http.StatusForbidden, "You don't have permission to view this form")
	}

	data := shared.BuildPageData(h.Config, "View Form")
	data.User = user
	data.Form = form
	return h.Renderer.Render(c, pages.Forms(data))
}
