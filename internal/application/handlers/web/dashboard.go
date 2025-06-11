package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
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

// handleDashboard handles the dashboard page request
func (h *DashboardHandler) handleDashboard(c echo.Context) error {
	ctx := c.Request().Context()
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok || userIDRaw == "" {
		// Let the error handler middleware handle the logging
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get user and forms in parallel
	userChan := make(chan *user.User, 1)
	formsChan := make(chan []*model.Form, 1)
	errChan := make(chan error, 2)

	go func() {
		userObj, err := h.UserService.GetUserByID(ctx, userIDRaw)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- userObj
	}()

	go func() {
		forms, err := h.FormService.GetUserForms(ctx, userIDRaw)
		if err != nil {
			errChan <- err
			return
		}
		formsChan <- forms
	}()

	// Wait for both operations to complete
	var userObj *user.User
	var forms []*model.Form
	var err error

	for i := 0; i < 2; i++ {
		select {
		case err = <-errChan:
			if errors.IsNotFound(err) {
				return response.WebErrorResponse(c, h.Renderer, http.StatusNotFound, "User not found")
			}
			return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to load dashboard data")
		case userObj = <-userChan:
		case forms = <-formsChan:
		}
	}

	data := shared.BuildPageData(h.Config, c, "Dashboard")
	data.User = userObj
	return h.Renderer.Render(c, pages.Dashboard(data, forms))
}

// handleFormView handles the form view page request
func (h *DashboardHandler) handleFormView(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get user and form in parallel
	userChan := make(chan *user.User, 1)
	formChan := make(chan *model.Form, 1)
	errChan := make(chan error, 2)

	go func() {
		userObj, err := h.UserService.GetUserByID(c.Request().Context(), userIDRaw)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- userObj
	}()

	go func() {
		form, err := h.FormService.GetForm(c.Request().Context(), formID)
		if err != nil {
			errChan <- err
			return
		}
		formChan <- form
	}()

	// Wait for both operations to complete
	var userObj *user.User
	var form *model.Form
	var err error

	for i := 0; i < 2; i++ {
		select {
		case err = <-errChan:
			if errors.IsNotFound(err) {
				return response.WebErrorResponse(c, h.Renderer, http.StatusNotFound, "Resource not found")
			}
			return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to load form data")
		case userObj = <-userChan:
		case form = <-formChan:
		}
	}

	// Verify form ownership
	if form.UserID != userIDRaw {
		// Log unauthorized access attempt with sanitized data
		h.Logger.Error("unauthorized form access attempt",
			"user_id", h.Logger.SanitizeField("user_id", userIDRaw),
			"form_id", h.Logger.SanitizeField("form_id", formID),
			"form_owner", h.Logger.SanitizeField("form_owner", form.UserID),
			"error_type", "authorization_error")
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to view this form")
	}

	data := shared.BuildPageData(h.Config, c, "View Form")
	data.User = userObj
	data.Form = form
	return h.Renderer.Render(c, pages.Forms(data))
}
