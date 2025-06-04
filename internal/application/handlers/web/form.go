package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type FormHandler struct {
	HandlerDeps
	FormService form.Service
}

func NewFormHandler(deps HandlerDeps, formService form.Service) *FormHandler {
	return &FormHandler{HandlerDeps: deps, FormService: formService}
}

func (h *FormHandler) Register(e *echo.Echo) {
	e.GET("/dashboard/forms/new", h.handleFormNew)
	e.POST("/dashboard/forms", h.handleFormCreate)
	e.GET("/dashboard/forms/:id/edit", h.handleFormEdit)
	e.PUT("/dashboard/forms/:id", h.handleFormUpdate)
	e.DELETE("/dashboard/forms/:id", h.handleFormDelete)
	e.GET("/dashboard/forms/:id/submissions", h.handleFormSubmissions)
}

// GET /dashboard/forms/new
func (h *FormHandler) handleFormNew(c echo.Context) error {
	data := shared.BuildPageData(h.Config, "New Form")
	return h.Renderer.Render(c, pages.NewForm(data))
}

// POST /dashboard/forms
func (h *FormHandler) handleFormCreate(c echo.Context) error {
	// TODO: Parse form data and create form
	return response.ErrorResponse(c, http.StatusNotImplemented, "Form creation not implemented yet")
}

// GET /dashboard/forms/:id/edit
func (h *FormHandler) handleFormEdit(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}
	f, err := h.FormService.GetForm(formID)
	if err != nil {
		h.Logger.Error("failed to get form", err)
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}
	data := shared.BuildPageData(h.Config, "Edit Form")
	data.Form = f
	return h.Renderer.Render(c, pages.EditForm(data))
}

// PUT /dashboard/forms/:id
func (h *FormHandler) handleFormUpdate(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}
	// TODO: Parse and update form details
	return response.ErrorResponse(c, http.StatusNotImplemented, "Form update not implemented yet")
}

// DELETE /dashboard/forms/:id
func (h *FormHandler) handleFormDelete(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}
	// TODO: Delete the form
	return response.ErrorResponse(c, http.StatusNotImplemented, "Form deletion not implemented yet")
}

// GET /dashboard/forms/:id/submissions
func (h *FormHandler) handleFormSubmissions(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}
	// TODO: Fetch submissions for the form
	var submissions []*model.FormSubmission // Placeholder
	data := shared.BuildPageData(h.Config, "Form Submissions")
	data.Submissions = submissions
	// TODO: Create a pages.FormSubmissions(data) template
	return h.Renderer.Render(c, pages.FormSubmissions(data))
}
