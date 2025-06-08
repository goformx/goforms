package web

import (
	"encoding/json"
	"errors"
	"fmt"
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
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}

	data := shared.BuildPageData(h.Config, "New Form")
	data.User = user
	if csrfToken, hasToken := c.Get("csrf").(string); hasToken {
		data.CSRFToken = csrfToken
	}
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
	schemaStr := c.FormValue("schema")

	// Log form creation attempt with raw schema
	h.Logger.Debug("attempting to create form",
		logging.StringField("title", title),
		logging.StringField("description", description),
		logging.StringField("raw_schema", schemaStr),
		logging.UintField("user_id", userID),
	)

	// Create a default schema if none provided
	var schema model.JSON
	if schemaStr == "" {
		schema = model.JSON{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{
					"type":  "string",
					"title": "Title",
				},
				"description": map[string]any{
					"type":  "string",
					"title": "Description",
				},
			},
			"required": []string{"title"},
		}
		h.Logger.Debug("using default schema",
			logging.StringField("schema", fmt.Sprintf("%+v", schema)),
		)
	} else {
		// Parse schema from form data
		if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
			h.Logger.Error("failed to parse form schema",
				logging.ErrorField("error", err),
				logging.StringField("raw_schema", schemaStr),
			)
			return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Invalid form schema format")
		}
	}

	// Create the form
	form, err := h.FormService.CreateForm(c.Request().Context(), userID, title, description, schema)
	if err != nil {
		h.Logger.Error("failed to create form",
			logging.ErrorField("error", err),
			logging.StringField("title", title),
			logging.StringField("description", description),
			logging.UintField("user_id", userID),
		)

		// Check for specific validation errors
		switch {
		case errors.Is(err, model.ErrFormTitleRequired):
			return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form title is required")
		case errors.Is(err, model.ErrFormSchemaRequired):
			return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form schema is required")
		default:
			return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to create form")
		}
	}

	// Log successful form creation
	h.Logger.Info("form created successfully",
		logging.StringField("form_id", form.ID),
		logging.StringField("title", form.Title),
		logging.UintField("user_id", form.UserID),
	)

	// Redirect to dashboard on success
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// GET /dashboard/forms/:id/edit
func (h *FormHandler) handleFormEdit(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
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
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}

	f, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form")
	}

	data := shared.BuildPageData(h.Config, "Edit Form")
	data.User = user
	data.Form = f
	if csrfToken, hasToken := c.Get("csrf").(string); hasToken {
		data.CSRFToken = csrfToken
	}
	return h.Renderer.Render(c, pages.EditForm(data))
}

// PUT /dashboard/forms/:id
func (h *FormHandler) handleFormUpdate(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}
	// TODO: Parse and update form details
	return response.WebErrorResponse(c, h.Renderer, http.StatusNotImplemented, "Form update not implemented yet")
}

// DELETE /dashboard/forms/:id
func (h *FormHandler) handleFormDelete(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(uint)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get form to verify ownership
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", logging.ErrorField("error", err))
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if form.UserID != userID {
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to delete this form")
	}

	// Delete the form
	if deleteErr := h.FormService.DeleteForm(c.Request().Context(), formID); deleteErr != nil {
		h.Logger.Error("failed to delete form", logging.ErrorField("error", deleteErr))
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to delete form")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Form deleted successfully",
	})
}

// GET /dashboard/forms/:id/submissions
func (h *FormHandler) handleFormSubmissions(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(uint)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get form to verify ownership
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", logging.ErrorField("error", err))
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if form.UserID != userID {
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to view these submissions")
	}

	// Get form submissions
	submissions, err := h.FormService.GetFormSubmissions(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form submissions", logging.ErrorField("error", err))
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form submissions")
	}

	// Get user object for the template
	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		h.Logger.Error("failed to get user (nil or error)", logging.ErrorField("error", err))
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}

	data := shared.BuildPageData(h.Config, "Form Submissions")
	data.User = user
	data.Form = form
	data.Submissions = submissions
	data.Content = pages.FormSubmissionsContent(data)
	return h.Renderer.Render(c, pages.FormSubmissions(data))
}

// GET /api/v1/forms/:id/schema
func (h *FormHandler) handleFormSchema(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "Form ID is required")
	}

	form, err := h.FormService.GetForm(c.Request().Context(), formID)
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
	form, getErr := h.FormService.GetForm(c.Request().Context(), formID)
	if getErr != nil {
		h.Logger.Error("failed to get form", logging.ErrorField("error", getErr))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	// Parse request body
	var schema model.JSON
	if decodeErr := json.NewDecoder(c.Request().Body).Decode(&schema); decodeErr != nil {
		h.Logger.Error("failed to decode request body", logging.ErrorField("error", decodeErr))
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	// Update form schema
	form.Schema = schema
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form); updateErr != nil {
		h.Logger.Error("failed to update form", logging.ErrorField("error", updateErr))
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update form")
	}

	// Return the updated schema
	return c.JSON(http.StatusOK, form.Schema)
}
