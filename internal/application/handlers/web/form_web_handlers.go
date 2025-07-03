// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
)

// Default CORS settings for forms
const (
	DefaultCorsMethods = "GET,POST,OPTIONS"
	DefaultCorsHeaders = "Content-Type,Accept,Origin"
)

// FormWebHandler handles web UI form operations
type FormWebHandler struct {
	*FormBaseHandler
	Sanitizer        sanitization.ServiceInterface
	RequestProcessor FormRequestProcessor
	ResponseBuilder  FormResponseBuilder
	ErrorHandler     FormErrorHandler
	FormService      *services.FormService
	AuthHelper       *AuthHelper
}

// NewFormWebHandler creates a new FormWebHandler instance
func NewFormWebHandler(
	base *BaseHandler,
	formService formdomain.Service,
	formValidator *validation.FormValidator,
	sanitizer sanitization.ServiceInterface,
) *FormWebHandler {
	// Create base handler
	formBaseHandler := NewFormBaseHandler(base, formService, formValidator)

	// Create dependencies
	requestProcessor := NewFormRequestProcessor(sanitizer, formValidator)
	responseBuilder := NewFormResponseBuilder()
	errorHandler := NewFormErrorHandler(responseBuilder)
	formServiceHandler := services.NewFormService(formService, base.Logger)
	authHelper := NewAuthHelper(formBaseHandler)

	return &FormWebHandler{
		FormBaseHandler:  formBaseHandler,
		Sanitizer:        sanitizer,
		RequestProcessor: requestProcessor,
		ResponseBuilder:  responseBuilder,
		ErrorHandler:     errorHandler,
		FormService:      formServiceHandler,
		AuthHelper:       authHelper,
	}
}

// RegisterRoutes registers all form-related routes
func (h *FormWebHandler) RegisterRoutes(e *echo.Echo, accessManager *access.Manager) {
	forms := e.Group(constants.PathForms)
	forms.Use(access.Middleware(accessManager, h.Logger))

	forms.GET("/new", h.handleNew)
	forms.POST("", h.handleCreate)
	forms.GET("/:id/edit", h.handleEdit)
	forms.POST("/:id/edit", h.handleUpdate)
	forms.PUT("/:id", h.handleUpdate)
	forms.DELETE("/:id", h.handleDelete)
	forms.GET("/:id/submissions", h.handleSubmissions)
	forms.GET("/:id/preview", h.handlePreview)

	// API routes with validation
	api := e.Group(constants.PathAPIV1)
	validationGroup := api.Group(constants.PathValidation)
	validationGroup.GET("/new-form", h.handleNewFormValidation)
}

// Register satisfies the Handler interface
func (h *FormWebHandler) Register(_ *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// Start initializes the form web handler.
func (h *FormWebHandler) Start(_ context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the form web handler.
func (h *FormWebHandler) Stop(_ context.Context) error {
	return nil // No cleanup needed
}

// HTTP handler methods for form operations
// These methods handle the actual HTTP request/response logic

// handleNew displays the new form creation page
func (h *FormWebHandler) handleNew(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	data := h.NewPageData(c, "New Form")
	data.SetUser(user)

	if renderErr := h.Renderer.Render(c, pages.NewForm(*data)); renderErr != nil {
		return fmt.Errorf("failed to render new form page: %w", renderErr)
	}

	return nil
}

// handleCreate processes form creation requests
func (h *FormWebHandler) handleCreate(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	// Process and validate request
	req, err := h.RequestProcessor.ProcessCreateRequest(c)
	if err != nil {
		return fmt.Errorf("handle error: %w", h.ErrorHandler.HandleError(c, err))
	}

	// Create form using business logic service
	form, err := h.FormService.CreateForm(c.Request().Context(), user.ID, req)
	if err != nil {
		return h.handleFormCreationError(c, err)
	}

	return fmt.Errorf("build success response: %w",
		h.ResponseBuilder.BuildSuccessResponse(c, "Form created successfully", map[string]any{
			"form_id": form.ID,
		}))
}

// handleEdit displays the form editing page
func (h *FormWebHandler) handleEdit(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	// Log form access for debugging
	h.FormService.LogFormAccess(form)

	data := h.NewPageData(c, "Edit Form").
		WithForm(form).
		WithFormBuilderAssetPath(h.AssetManager.AssetPath("src/js/pages/form-builder.ts"))

	data.SetUser(user)

	if renderErr := h.Renderer.Render(c, pages.EditForm(*data, form)); renderErr != nil {
		return fmt.Errorf("failed to render edit form page: %w", renderErr)
	}

	return nil
}

// handleUpdate processes form update requests
func (h *FormWebHandler) handleUpdate(c echo.Context) error {
	_, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	// Process and validate request
	req, err := h.RequestProcessor.ProcessUpdateRequest(c)
	if err != nil {
		return fmt.Errorf("handle error: %w", h.ErrorHandler.HandleError(c, err))
	}

	// Update form using business logic service
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form, req); updateErr != nil {
		h.Logger.Error("failed to update form", "error", updateErr)

		return h.HandleError(c, updateErr, "Failed to update form")
	}

	return fmt.Errorf("build success response: %w",
		h.ResponseBuilder.BuildSuccessResponse(c, "Form updated successfully", map[string]any{
			"form_id": form.ID,
		}))
}

// handleDelete processes form deletion requests
func (h *FormWebHandler) handleDelete(c echo.Context) error {
	_, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	if deleteErr := h.FormService.DeleteForm(c.Request().Context(), form.ID); deleteErr != nil {
		h.Logger.Error("failed to delete form", "error", deleteErr)

		return h.HandleError(c, deleteErr, "Failed to delete form")
	}

	return fmt.Errorf("no content response: %w", c.NoContent(constants.StatusNoContent))
}

// handleSubmissions displays form submissions
func (h *FormWebHandler) handleSubmissions(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	submissions, err := h.FormService.GetFormSubmissions(c.Request().Context(), form.ID)
	if err != nil {
		h.Logger.Error("failed to get form submissions", "error", err)

		return h.HandleError(c, err, "Failed to get form submissions")
	}

	data := h.NewPageData(c, "Form Submissions").
		WithForm(form).
		WithSubmissions(submissions)

	data.SetUser(user)

	if renderErr := h.Renderer.Render(c, pages.FormSubmissions(*data)); renderErr != nil {
		return fmt.Errorf("failed to render form submissions page: %w", renderErr)
	}

	return nil
}

// handlePreview handles the form preview page request
func (h *FormWebHandler) handlePreview(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return h.HandleForbidden(c, "User not authenticated")
	}

	formID := c.Param("id")
	if formID == "" {
		return h.HandleError(c, nil, "Form ID is required")
	}

	// Fetch form data using the domain service
	form, err := h.FormBaseHandler.FormService.GetForm(c.Request().Context(), formID)
	if err != nil || form == nil {
		h.Logger.Warn("form preview access attempt failed",
			"user_id", h.Logger.SanitizeField("user_id", userID),
			"form_id_length", len(formID),
			"error_type", "form_not_found")

		return h.HandleNotFound(c, "Form not found")
	}

	// Verify form ownership
	if form.UserID != userID {
		h.Logger.Warn("unauthorized form preview access attempt",
			"user_id", h.Logger.SanitizeField("user_id", userID),
			"form_id_length", len(formID),
			"form_owner", h.Logger.SanitizeField("form_owner", form.UserID),
			"error_type", "authorization_error")

		return h.HandleForbidden(c, "You don't have permission to preview this form")
	}

	h.Logger.Debug("form preview accessed successfully",
		"user_id", h.Logger.SanitizeField("user_id", userID),
		"form_id_length", len(formID),
		"form_title", h.Logger.SanitizeField("form_title", form.Title))

	// Build page data using the new API
	data := h.NewPageData(c, "Form Preview").
		WithForm(form).
		WithFormPreviewAssetPath(h.AssetManager.AssetPath("src/js/pages/form-preview.ts"))

	// Render form preview template
	if renderErr := h.Renderer.Render(c, pages.FormPreview(*data, form)); renderErr != nil {
		return fmt.Errorf("failed to render form preview page: %w", renderErr)
	}

	return nil
}

// handleNewFormValidation returns the validation schema for the new form
func (h *FormWebHandler) handleNewFormValidation(c echo.Context) error {
	schema := map[string]any{
		"title": map[string]any{
			"type":    "required",
			"message": "Form title is required",
		},
	}

	return response.Success(c, schema)
}

// handleFormCreationError handles form creation errors
func (h *FormWebHandler) handleFormCreationError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrFormTitleRequired):
		return fmt.Errorf("build error response: %w",
			h.ResponseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form title is required"))
	case errors.Is(err, model.ErrFormSchemaRequired):
		return fmt.Errorf("build error response: %w",
			h.ResponseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form schema is required"))
	default:
		return fmt.Errorf("build error response: %w",
			h.ResponseBuilder.BuildErrorResponse(c, http.StatusInternalServerError, "Failed to create form"))
	}
}
