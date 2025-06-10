package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
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
	if err := deps.Validate(); err != nil {
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
	data := shared.BuildPageData(h.Config, c, "Home")
	if err := h.Renderer.Render(c, pages.Home(data)); err != nil {
		data.Message = &shared.Message{
			Type: "error",
			Text: err.Error(),
		}
		return pages.Error(data).Render(c.Request().Context(), c.Response().Writer)
	}
	return nil
}

// handleDashboard handles the dashboard page request
func (h *WebHandler) handleDashboard(c echo.Context) error {
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
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}
	data := shared.BuildPageData(h.Config, c, "Dashboard")
	data.User = user

	// Get user's forms
	forms, err := h.FormService.GetUserForms(c.Request().Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get user forms",
			"operation", "handle_dashboard",
			"user_id", userID,
			"error", err,
		)
		data.Message = &shared.Message{
			Type: "error",
			Text: "Failed to load forms. Please try again later.",
		}
		return h.Renderer.Render(c, pages.Error(data))
	}

	h.Logger.Debug("rendering dashboard page")
	return pages.Dashboard(data, forms).Render(c.Request().Context(), c.Response().Writer)
}

// handleFormView handles the form view page request
func (h *WebHandler) handleFormView(c echo.Context) error {
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
