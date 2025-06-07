package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware"
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
	h.Logger.Debug("dashboard request",
		logging.StringField("path", c.Request().URL.Path),
		logging.StringField("method", c.Request().Method),
	)

	// Get session from context
	session, ok := c.Get(middleware.SessionKey).(*middleware.Session)
	if !ok {
		h.Logger.Error("no session found in context",
			logging.StringField("path", c.Request().URL.Path),
			logging.StringField("method", c.Request().Method),
		)
		return response.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated")
	}

	h.Logger.Debug("session found in context",
		logging.UintField("user_id", session.UserID),
		logging.StringField("email", session.Email),
		logging.StringField("role", session.Role),
	)

	// Get user data
	user, err := h.UserService.GetUserByID(c.Request().Context(), session.UserID)
	if err != nil {
		h.Logger.Error("failed to get user data",
			logging.ErrorField("error", err),
			logging.UintField("user_id", session.UserID),
		)
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user data")
	}

	// Get user's forms
	forms, err := h.BaseHandler.formService.GetUserForms(session.UserID)
	if err != nil {
		h.Logger.Error("failed to get user's forms",
			logging.ErrorField("error", err),
			logging.UintField("user_id", session.UserID),
		)
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

	form, err := h.BaseHandler.formService.GetForm(formID)
	if err != nil {
		h.Logger.Error("failed to get form", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	data := shared.BuildPageData(h.Config, form.Title)
	data.Form = form
	return h.Renderer.Render(c, pages.Forms(data))
}
