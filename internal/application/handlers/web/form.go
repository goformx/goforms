package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
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
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get form data
	title := c.FormValue("title")
	description := c.FormValue("description")

	// Log form creation attempt
	h.Logger.Debug("attempting to create form",
		"title", title,
		"description", description,
		"user_id", userID,
	)

	// Create a valid initial schema
	schema := model.JSON{
		"type":       "object",
		"components": []any{},
	}

	// Create the form
	form := model.NewForm(userID, title, description, schema)
	err := h.FormService.CreateForm(c.Request().Context(), userID, form)
	if err != nil {
		h.Logger.Error("failed to create form",
			"error", err,
			"title", title,
			"description", description,
			"user_id", userID,
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
		"form_id", form.ID,
		"title", form.Title,
		"user_id", form.UserID,
	)

	// Redirect to the form edit page
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/dashboard/forms/%s/edit", form.ID))
}

// GET /dashboard/forms/:id/edit
func (h *FormHandler) handleFormEdit(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		h.Logger.Error("form ID is required",
			"operation", "handle_form_edit",
			"error_type", "validation_error",
		)
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok {
		h.Logger.Error("user ID not found in session",
			"operation", "handle_form_edit",
			"error_type", "session_error",
		)
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get user object
	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		h.Logger.Error("failed to get user (nil or error)",
			"error", err,
			"user_id", userID,
			"operation", "handle_form_edit",
			"error_type", "user_service_error",
		)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}

	h.Logger.Debug("attempting to get form",
		"form_id", formID,
	)

	f, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if f.UserID != userID {
		h.Logger.Error("form ownership verification failed",
			"form_id", formID,
			"user_id", userID,
			"form_user_id", f.UserID,
			"operation", "handle_form_edit",
			"error_type", "authorization_error",
		)
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to edit this form")
	}

	data := shared.BuildPageData(h.Config, "Edit Form")
	data.User = user
	data.Form = f
	if csrfToken, hasToken := c.Get("csrf").(string); hasToken {
		data.CSRFToken = csrfToken
	}

	h.Logger.Debug("form edit page rendered successfully",
		"form_id", formID,
		"user_id", userID,
		"operation", "handle_form_edit",
	)

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
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get form to verify ownership
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if form.UserID != userID {
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to delete this form")
	}

	// Delete the form
	if deleteErr := h.FormService.DeleteForm(c.Request().Context(), userID, formID); deleteErr != nil {
		h.Logger.Error("failed to delete form", "error", deleteErr)
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

	// Get form to verify ownership
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if form.UserID != userID {
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to view submissions for this form")
	}

	// Get form submissions
	submissions, err := h.FormService.GetFormSubmissions(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form submissions", "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form submissions")
	}

	data := shared.BuildPageData(h.Config, "Form Submissions")
	data.User = user
	data.Form = form
	data.Submissions = submissions
	if csrfToken, hasToken := c.Get("csrf").(string); hasToken {
		data.CSRFToken = csrfToken
	}

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
		h.Logger.Error("failed to get form schema", "error", err)
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
		h.Logger.Error("failed to get form", "error", getErr)
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to get form")
	}

	// Parse request body
	var schema model.JSON
	if decodeErr := json.NewDecoder(c.Request().Body).Decode(&schema); decodeErr != nil {
		h.Logger.Error("failed to decode request body", "error", decodeErr)
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(string)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Update form schema
	form.Schema = schema
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), userID, form); updateErr != nil {
		h.Logger.Error("failed to update form", "error", updateErr)
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update form")
	}

	// Return the updated schema
	return c.JSON(http.StatusOK, form.Schema)
}
