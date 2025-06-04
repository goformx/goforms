package web

import (
	"encoding/json"
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// PageHandler handles page-related requests
type PageHandler struct {
	HandlerDeps
	FormService form.Service // for now, keep this for direct access
}

// NewPageHandler creates a new page handler using HandlerDeps
func NewPageHandler(deps HandlerDeps, formService form.Service) (*PageHandler, error) {
	if err := deps.Validate("Logger"); err != nil {
		return nil, err
	}
	return &PageHandler{HandlerDeps: deps, FormService: formService}, nil
}

// Register registers the page routes
func (h *PageHandler) Register(e *echo.Echo) {
	e.GET("/pages", h.handlePages)
	e.GET("/pages/:id", h.handlePageView)
	e.POST("/pages", h.handlePageCreate)
	e.PUT("/pages/:id", h.handlePageUpdate)
	e.DELETE("/pages/:id", h.handlePageDelete)
}

// handlePages handles the pages list request
func (h *PageHandler) handlePages(c echo.Context) error {
	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(uint)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get user's forms
	forms, err := h.FormService.GetUserForms(userID)
	if err != nil {
		h.Logger.Error("failed to get user forms", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get forms")
	}

	data := shared.BuildPageData(h.Config, "Pages")
	data.Forms = forms
	return pages.PagesList(data).Render(c.Request().Context(), c.Response().Writer)
}

// handlePageView handles the page view request
func (h *PageHandler) handlePageView(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}

	formData, err := h.FormService.GetForm(formID)
	if err != nil {
		h.Logger.Error("failed to get form", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	data := shared.BuildPageData(h.Config, formData.Title)
	data.Form = formData
	return pages.PageView(data).Render(c.Request().Context(), c.Response().Writer)
}

// handlePageCreate handles the page creation request
func (h *PageHandler) handlePageCreate(c echo.Context) error {
	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(uint)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Parse schema JSON
	var schema form.JSON
	schemaJSON := c.FormValue("schema")
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid schema JSON")
	}

	_, err := h.FormService.CreateForm(userID, c.FormValue("title"), c.FormValue("description"), schema)
	if err != nil {
		h.Logger.Error("failed to create form", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create form")
	}

	return c.Redirect(http.StatusSeeOther, "/pages")
}

// handlePageUpdate handles the page update request
func (h *PageHandler) handlePageUpdate(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}

	// Get existing form
	existingForm, getErr := h.FormService.GetForm(formID)
	if getErr != nil {
		h.Logger.Error("failed to get form", logging.ErrorField("error", getErr))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	// Parse schema JSON
	var schema form.JSON
	schemaJSON := c.FormValue("schema")
	if unmarshalErr := json.Unmarshal([]byte(schemaJSON), &schema); unmarshalErr != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid schema JSON")
	}

	// Update form fields
	existingForm.Title = c.FormValue("title")
	existingForm.Description = c.FormValue("description")
	existingForm.Schema = schema

	if updateErr := h.FormService.UpdateForm(existingForm); updateErr != nil {
		h.Logger.Error("failed to update form", logging.ErrorField("error", updateErr))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update form")
	}

	return c.Redirect(http.StatusSeeOther, "/pages")
}

// handlePageDelete handles the page deletion request
func (h *PageHandler) handlePageDelete(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}

	if err := h.FormService.DeleteForm(formID); err != nil {
		h.Logger.Error("failed to delete form", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete form")
	}

	return c.Redirect(http.StatusSeeOther, "/pages")
}
