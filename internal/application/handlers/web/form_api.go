package web

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// FormAPIHandler handles API form operations
type FormAPIHandler struct {
	*FormBaseHandler
	AccessManager          *access.Manager
	RequestProcessor       FormRequestProcessor
	ResponseBuilder        FormResponseBuilder
	ErrorHandler           FormErrorHandler
	ComprehensiveValidator *validation.ComprehensiveValidator
}

// NewFormAPIHandler creates a new FormAPIHandler.
func NewFormAPIHandler(
	base *BaseHandler,
	formService formdomain.Service,
	accessManager *access.Manager,
	formValidator *validation.FormValidator,
	sanitizer sanitization.ServiceInterface,
) *FormAPIHandler {
	// Create dependencies
	requestProcessor := NewFormRequestProcessor(sanitizer, formValidator)
	responseBuilder := NewFormResponseBuilder()
	errorHandler := NewFormErrorHandler(responseBuilder)
	comprehensiveValidator := validation.NewComprehensiveValidator()

	return &FormAPIHandler{
		FormBaseHandler:        NewFormBaseHandler(base, formService, formValidator),
		AccessManager:          accessManager,
		RequestProcessor:       requestProcessor,
		ResponseBuilder:        responseBuilder,
		ErrorHandler:           errorHandler,
		ComprehensiveValidator: comprehensiveValidator,
	}
}

// RegisterRoutes registers API routes for forms.
func (h *FormAPIHandler) RegisterRoutes(e *echo.Echo) {
	api := e.Group(constants.PathAPIv1)
	formsAPI := api.Group(constants.PathForms)

	// Register authenticated routes
	h.RegisterAuthenticatedRoutes(formsAPI)

	// Register public routes
	h.RegisterPublicRoutes(formsAPI)
}

// RegisterAuthenticatedRoutes registers routes that require authentication
func (h *FormAPIHandler) RegisterAuthenticatedRoutes(formsAPI *echo.Group) {
	// Apply authentication middleware
	formsAPI.Use(access.Middleware(h.AccessManager, h.Logger))

	// Authenticated routes
	formsAPI.GET("/:id/schema", h.handleFormSchema)
	formsAPI.PUT("/:id/schema", h.handleFormSchemaUpdate)
}

// RegisterPublicRoutes registers routes that don't require authentication
func (h *FormAPIHandler) RegisterPublicRoutes(formsAPI *echo.Group) {
	// Public routes (no authentication required)
	// These are for embedded forms on external websites
	formsAPI.GET("/:id/schema", h.handleFormSchema)
	formsAPI.GET("/:id/validation", h.handleFormValidationSchema)
	formsAPI.POST("/:id/submit", h.handleFormSubmit)
}

// Register registers the FormAPIHandler with the Echo instance.
func (h *FormAPIHandler) Register(_ *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// GET /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchema(c echo.Context) error {
	form, err := h.GetFormByID(c)
	if err != nil {
		return h.HandleError(c, err, "Failed to get form schema")
	}

	// Check if form is nil (should not happen with proper error handling, but safety check)
	if form == nil {
		h.Logger.Error("form is nil after GetFormByID", "form_id", c.Param("id"))
		return fmt.Errorf("handle form schema: %w", h.ErrorHandler.HandleFormNotFoundError(c, ""))
	}

	return fmt.Errorf("build schema response: %w", h.ResponseBuilder.BuildSchemaResponse(c, form.Schema))
}

// GET /api/v1/forms/:id/validation
func (h *FormAPIHandler) handleFormValidationSchema(c echo.Context) error {
	form, err := h.GetFormByID(c)
	if err != nil {
		return h.HandleError(c, err, "Failed to get form for validation schema")
	}

	// Check if form is nil (should not happen with proper error handling, but safety check)
	if form == nil {
		h.Logger.Error("form is nil after GetFormByID", "form_id", c.Param("id"))
		return fmt.Errorf("handle form validation schema: %w", h.ErrorHandler.HandleFormNotFoundError(c, ""))
	}

	// Check if form schema is nil or empty
	if form.Schema == nil {
		h.Logger.Warn("form schema is nil", "form_id", form.ID)
		return c.JSON(constants.StatusOK, map[string]any{})
	}

	// Generate client-side validation rules from form schema
	clientValidation, err := h.ComprehensiveValidator.GenerateClientValidation(form.Schema)
	if err != nil {
		h.Logger.Error("failed to generate client validation schema", "error", err, "form_id", form.ID)
		return fmt.Errorf("handle schema error: %w", h.ErrorHandler.HandleSchemaError(c, err))
	}

	return c.JSON(constants.StatusOK, clientValidation)
}

// PUT /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchemaUpdate(c echo.Context) error {
	_, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return fmt.Errorf("handle ownership error: %w", h.ErrorHandler.HandleOwnershipError(c, err))
	}

	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return h.HandleError(c, err, "Unauthorized or form not found")
	}

	// Process and validate schema update request
	schema, err := h.RequestProcessor.ProcessSchemaUpdateRequest(c)
	if err != nil {
		return fmt.Errorf("handle schema error: %w", h.ErrorHandler.HandleSchemaError(c, err))
	}

	// Update form schema
	form.Schema = schema
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form); updateErr != nil {
		h.Logger.Error("failed to update form schema", "error", updateErr)
		return fmt.Errorf("handle schema update error: %w", h.ErrorHandler.HandleSchemaError(c, updateErr))
	}

	return fmt.Errorf("build schema response: %w", h.ResponseBuilder.BuildSchemaResponse(c, form.Schema))
}

// POST /api/v1/forms/:id/submit
func (h *FormAPIHandler) handleFormSubmit(c echo.Context) error {
	form, err := h.GetFormByID(c)
	if err != nil {
		return h.HandleError(c, err, "Failed to get form for submission")
	}

	// Check if form is nil (should not happen with proper error handling, but safety check)
	if form == nil {
		h.Logger.Error("form is nil after GetFormByID", "form_id", c.Param("id"))
		return fmt.Errorf("handle form submit: %w", h.ErrorHandler.HandleFormNotFoundError(c, ""))
	}

	// Check if form schema is nil or empty
	if form.Schema == nil {
		h.Logger.Warn("form schema is nil", "form_id", form.ID)
		return fmt.Errorf("handle submission error: %w", h.ErrorHandler.HandleSchemaError(c, errors.New("form schema is required")))
	}

	// Process and validate submission request
	submissionData, err := h.RequestProcessor.ProcessSubmissionRequest(c)
	if err != nil {
		return fmt.Errorf("handle submission error: %w", h.ErrorHandler.HandleSubmissionError(c, err))
	}

	// Validate submission against form schema
	validationResult := h.ComprehensiveValidator.ValidateForm(form.Schema, submissionData)
	if !validationResult.IsValid {
		return fmt.Errorf("build multiple error response: %w", h.ResponseBuilder.BuildMultipleErrorResponse(c, validationResult.Errors))
	}

	// Create submission
	submission := &model.FormSubmission{
		FormID:      form.ID,
		Data:        submissionData,
		SubmittedAt: time.Now(),
		Status:      model.SubmissionStatusPending,
	}

	// Submit form
	err = h.FormService.SubmitForm(c.Request().Context(), submission)
	if err != nil {
		return fmt.Errorf("handle submission error: %w", h.ErrorHandler.HandleSubmissionError(c, err))
	}

	return fmt.Errorf("build submission response: %w", h.ResponseBuilder.BuildSubmissionResponse(c, submission))
}

// Start initializes the form API handler.
// This is called during application startup.
func (h *FormAPIHandler) Start(_ context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the form API handler.
// This is called during application shutdown.
func (h *FormAPIHandler) Stop(_ context.Context) error {
	return nil // No cleanup needed
}
