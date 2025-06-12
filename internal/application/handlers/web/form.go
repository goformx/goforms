package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware/access"
	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/response"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
	"github.com/mrz1836/go-sanitize"
)

type FormHandler struct {
	HandlerDeps
	FormService   formdomain.Service
	AccessManager *access.AccessManager
}

// NewFormHandler creates a new form handler
func NewFormHandler(
	deps HandlerDeps,
	formService formdomain.Service,
	accessManager *access.AccessManager,
) *FormHandler {
	return &FormHandler{
		HandlerDeps:   deps,
		FormService:   formService,
		AccessManager: accessManager,
	}
}

func (h *FormHandler) Register(e *echo.Echo) {
	// Web routes with access control
	forms := e.Group("/forms")
	forms.Use(access.Middleware(h.AccessManager, h.Logger))
	forms.GET("/new", h.handleFormNew)
	forms.POST("", h.handleFormCreate)
	forms.GET("/:id/edit", h.handleFormEdit)
	forms.PUT("/:id", h.handleFormUpdate)
	forms.DELETE("/:id", h.handleFormDelete)
	forms.GET("/:id/submissions", h.handleFormSubmissions)

	// API routes with access control
	api := e.Group("/api/v1")
	formsAPI := api.Group("/forms")
	formsAPI.Use(access.Middleware(h.AccessManager, h.Logger))
	formsAPI.GET("/:id/schema", h.handleFormSchema)
	formsAPI.PUT("/:id/schema", h.handleFormSchemaUpdate)
}

// GET /forms/new
func (h *FormHandler) handleFormNew(c echo.Context) error {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get user object
	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		h.Logger.Error("failed to get user", "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}

	data := shared.BuildPageData(h.Config, c, "New Form")
	data.User = user
	return h.Renderer.Render(c, pages.NewForm(data))
}

// POST /forms
func (h *FormHandler) handleFormCreate(c echo.Context) error {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get and sanitize form data
	title := sanitize.XSS(c.FormValue("title"))
	description := sanitize.XSS(c.FormValue("description"))

	// Create a valid initial schema
	schema := model.JSON{
		"type":       "object",
		"components": []any{},
	}

	// Create the form
	form := model.NewForm(userID, title, description, schema)
	err := h.FormService.CreateForm(c.Request().Context(), userID, form)
	if err != nil {
		h.Logger.Error("failed to create form", "error", err)

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

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/forms/%s/edit", form.ID))
}

// handleFormEdit handles the form edit page
func (h *FormHandler) handleFormEdit(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Get user ID from session
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Get user object
	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		h.Logger.Error("failed to get user", "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}

	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form",
			"form_id", formID,
			"error", err,
		)
		return response.WebErrorResponse(c, h.Renderer, http.StatusNotFound, "Form not found")
	}

	// Verify form ownership
	if form.UserID != userID {
		h.Logger.Error("form ownership verification failed", "form_id", formID, "user_id", userID)
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to edit this form")
	}

	data := shared.BuildPageData(h.Config, c, "Edit Form")
	data.User = user
	data.Form = form

	return pages.EditForm(data, form).Render(c.Request().Context(), c.Response().Writer)
}

// PUT /forms/:id
func (h *FormHandler) handleFormUpdate(c echo.Context) error {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Get form to verify ownership
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form")
	}

	// Verify form ownership
	if form.UserID != userID {
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, "You don't have permission to update this form")
	}

	// TODO: Parse and update form details
	return response.WebErrorResponse(c, h.Renderer, http.StatusNotImplemented, "Form update not implemented yet")
}

// DELETE /forms/:id
func (h *FormHandler) handleFormDelete(c echo.Context) error {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

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

// GET /forms/:id/submissions
func (h *FormHandler) handleFormSubmissions(c echo.Context) error {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	formID := c.Param("id")
	if formID == "" {
		return response.WebErrorResponse(c, h.Renderer, http.StatusBadRequest, "Form ID is required")
	}

	// Get user object
	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		h.Logger.Error("failed to get user", "error", err)
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
		h.Logger.Error("unauthorized form access attempt",
			"user_id", h.Logger.SanitizeField("user_id", userID),
			"form_id", h.Logger.SanitizeField("form_id", formID),
			"form_owner", h.Logger.SanitizeField("form_owner", form.UserID),
			"error_type", "authorization_error")
		return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden,
			"You don't have permission to view submissions for this form")
	}

	// Get form submissions
	submissions, err := h.FormService.GetFormSubmissions(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form submissions", "error", err)
		return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get form submissions")
	}

	data := shared.BuildPageData(h.Config, c, "Form Submissions")
	data.User = user
	data.Form = form
	data.Submissions = submissions

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
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

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

	// Verify form ownership
	if form.UserID != userID {
		return response.ErrorResponse(c, http.StatusForbidden, "You don't have permission to update this form")
	}

	// Parse request body
	var schema model.JSON
	if decodeErr := json.NewDecoder(c.Request().Body).Decode(&schema); decodeErr != nil {
		h.Logger.Error("failed to decode request body", "error", decodeErr)
		return response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	// Update form schema
	form.Schema = schema
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), userID, form); updateErr != nil {
		h.Logger.Error("failed to update form", "error", updateErr)
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update form")
	}

	// Return the updated schema
	return c.JSON(http.StatusOK, form.Schema)
}

// Start initializes the form handler.
// This is called during application startup.
func (h *FormHandler) Start(ctx context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the form handler.
// This is called during application shutdown.
func (h *FormHandler) Stop(ctx context.Context) error {
	return nil // No cleanup needed
}
