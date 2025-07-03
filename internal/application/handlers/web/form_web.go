// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
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
	FormService      *FormService
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
	formServiceHandler := NewFormService(formService, base.Logger)
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
	return h.Renderer.Render(c, pages.FormPreview(*data, form))
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
