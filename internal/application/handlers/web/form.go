package web

import (
	"encoding/json"
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type FormHandler struct {
	HandlerDeps
	FormService formdomain.Service
}

func NewFormHandler(deps HandlerDeps, formService formdomain.Service) *FormHandler {
	return &FormHandler{HandlerDeps: deps, FormService: formService}
}

func (h *FormHandler) Register(e *echo.Echo) {
	// Web routes
	e.GET("/dashboard/forms/new", h.handleFormNew)
	e.POST("/dashboard/forms", h.handleFormCreate)
	e.GET("/dashboard/forms/:id/edit", h.handleFormEdit)
	e.PUT("/dashboard/forms/:id", h.handleFormUpdate)
	e.DELETE("/dashboard/forms/:id", h.handleFormDelete)
	e.GET("/dashboard/forms/:id/submissions", h.handleFormSubmissions)

	// API routes
	api := e.Group("/api/v1")
	forms := api.Group("/forms")
	forms.GET("/:id/schema", h.handleFormSchema)
	forms.PUT("/:id/schema", h.handleFormSchemaUpdate)
}

// GET /dashboard/forms/new
func (h *FormHandler) handleFormNew(c echo.Context) error {
	data := shared.BuildPageData(h.Config, "New Form")
	return h.Renderer.Render(c, pages.NewForm(data))
}

// POST /dashboard/forms
func (h *FormHandler) handleFormCreate(c echo.Context) error {
	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(uint)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get form data
	title := c.FormValue("title")
	description := c.FormValue("description")

	// Create an empty schema for now
	schema := formdomain.JSON{
		"components": []any{},
	}

	// Create the form
	_, err := h.FormService.CreateForm(userID, title, description, schema)
	if err != nil {
		h.Logger.Error("failed to create form", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create form")
	}

	// Redirect to dashboard on success
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// GET /dashboard/forms/:id/edit
func (h *FormHandler) handleFormEdit(c echo.Context) error {
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

	f, err := h.FormService.GetForm(formID)
	if err != nil {
		h.Logger.Error("failed to get form", err)
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	data := shared.BuildPageData(h.Config, "Edit Form")
	data.User = user
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

// GET /api/v1/forms/:id/schema
func (h *FormHandler) handleFormSchema(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}

	form, err := h.FormService.GetForm(formID)
	if err != nil {
		h.Logger.Error("failed to get form schema", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form schema")
	}

	return c.JSON(http.StatusOK, form.Schema)
}

// PUT /api/v1/forms/:id/schema
func (h *FormHandler) handleFormSchemaUpdate(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}

	// Get existing form
	form, err := h.FormService.GetForm(formID)
	if err != nil {
		h.Logger.Error("failed to get form", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	// Parse schema JSON
	var schema formdomain.JSON
	if err := json.NewDecoder(c.Request().Body).Decode(&schema); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid schema JSON")
	}

	// Update form schema
	form.Schema = schema
	if err := h.FormService.UpdateForm(form); err != nil {
		h.Logger.Error("failed to update form schema", logging.ErrorField("error", err))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update form schema")
	}

	// Return the updated schema
	return c.JSON(http.StatusOK, schema)
}
