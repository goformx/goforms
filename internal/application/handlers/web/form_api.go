package web

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/response"
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
	formsAPI.GET("", h.handleListForms)
	formsAPI.GET("/:id", h.handleGetForm)
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

// GET /api/v1/forms
func (h *FormAPIHandler) handleListForms(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return h.HandleForbidden(c, "User not authenticated")
	}

	// Get forms for the user
	forms, err := h.FormService.ListForms(c.Request().Context(), userID)
	if err != nil {
		h.Logger.Error("failed to list forms", "error", err)

		return h.HandleError(c, err, "Failed to list forms")
	}

	h.Logger.Debug("forms listed successfully",
		"user_id", h.Logger.SanitizeField("user_id", userID),
		"form_count", len(forms))

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildFormListResponse(c, forms); respErr != nil {
		h.Logger.Error("failed to build form list response", "error", respErr)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// GET /api/v1/forms/:id
func (h *FormAPIHandler) handleGetForm(c echo.Context) error {
	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildFormResponse(c, form); respErr != nil {
		h.Logger.Error("failed to build form response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// GET /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchema(c echo.Context) error {
	form, err := h.getFormOrError(c)
	if err != nil {
		return err
	}

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildSchemaResponse(c, form.Schema); respErr != nil {
		h.Logger.Error("failed to build schema response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// GET /api/v1/forms/:id/validation
func (h *FormAPIHandler) handleFormValidationSchema(c echo.Context) error {
	form, err := h.getFormOrError(c)
	if err != nil {
		return err
	}

	if validationErr := h.validateFormSchema(c, form); validationErr != nil {
		return validationErr
	}

	// Generate client-side validation rules from form schema
	clientValidation, err := h.ComprehensiveValidator.GenerateClientValidation(form.Schema)
	if err != nil {
		h.Logger.Error("failed to generate client validation schema", "error", err, "form_id", form.ID)

		return h.wrapError("handle schema error", h.ErrorHandler.HandleSchemaError(c, err))
	}

	return response.Success(c, clientValidation)
}

// PUT /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchemaUpdate(c echo.Context) error {
	_, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return h.wrapError("handle ownership error", h.ErrorHandler.HandleOwnershipError(c, err))
	}

	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	// Process and validate schema update request
	schema, err := h.RequestProcessor.ProcessSchemaUpdateRequest(c)
	if err != nil {
		return h.wrapError("handle schema error", h.ErrorHandler.HandleSchemaError(c, err))
	}

	// Update form schema
	if updateErr := h.updateFormSchema(c, form, schema); updateErr != nil {
		return updateErr
	}

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildSchemaResponse(c, form.Schema); respErr != nil {
		h.Logger.Error("failed to build schema response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// POST /api/v1/forms/:id/submit
func (h *FormAPIHandler) handleFormSubmit(c echo.Context) error {
	formID := c.Param("id")
	h.logFormSubmissionRequest(c, formID)

	form, err := h.getFormOrError(c)
	if err != nil {
		return err
	}

	if validationErr := h.validateFormSchema(c, form); validationErr != nil {
		return validationErr
	}

	submissionData, err := h.processSubmissionRequest(c, form.ID)
	if err != nil {
		return err
	}

	if validationDataErr := h.validateSubmissionData(c, form, submissionData); validationDataErr != nil {
		return validationDataErr
	}

	submission, err := h.createAndSubmitForm(c, form, submissionData)
	if err != nil {
		return err
	}

	h.Logger.Info("Form submitted successfully", "form_id", form.ID, "submission_id", submission.ID)

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildSubmissionResponse(c, submission); respErr != nil {
		h.Logger.Error(
			"failed to build submission response",
			"error", respErr,
			"form_id", form.ID,
			"submission_id", submission.ID,
		)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
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

// Helper methods to reduce code duplication and improve SRP

// getFormOrError retrieves a form by ID and handles common error cases
func (h *FormAPIHandler) getFormOrError(c echo.Context) (*model.Form, error) {
	form, err := h.GetFormByID(c)
	if err != nil {
		return nil, h.HandleError(c, err, "Failed to get form")
	}

	if form == nil {
		h.Logger.Error("form is nil after GetFormByID", "form_id", c.Param("id"))

		return nil, h.wrapError("handle form not found", h.ErrorHandler.HandleFormNotFoundError(c, ""))
	}

	return form, nil
}

// getFormWithOwnershipOrError retrieves a form with ownership verification
func (h *FormAPIHandler) getFormWithOwnershipOrError(c echo.Context) (*model.Form, error) {
	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return nil, h.HandleError(c, err, "Failed to get form")
	}

	if form == nil {
		h.Logger.Error("form is nil after GetFormWithOwnership", "form_id", c.Param("id"))

		return nil, h.wrapError("handle form not found", h.ErrorHandler.HandleFormNotFoundError(c, ""))
	}

	return form, nil
}

// validateFormSchema validates that form schema exists
func (h *FormAPIHandler) validateFormSchema(c echo.Context, form *model.Form) error {
	if form.Schema == nil {
		h.Logger.Warn("form schema is nil", "form_id", form.ID)

		return h.wrapError("handle submission error",
			h.ErrorHandler.HandleSchemaError(c, errors.New("form schema is required")))
	}

	return nil
}

// updateFormSchema updates the form schema in the database
func (h *FormAPIHandler) updateFormSchema(c echo.Context, form *model.Form, schema model.JSON) error {
	form.Schema = schema
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form); updateErr != nil {
		h.Logger.Error("failed to update form schema", "error", updateErr)

		return h.wrapError("handle schema update error", h.ErrorHandler.HandleSchemaError(c, updateErr))
	}

	return nil
}

// logFormSubmissionRequest logs the initial form submission request
func (h *FormAPIHandler) logFormSubmissionRequest(c echo.Context, formID string) {
	h.Logger.Debug("Form submission request received",
		"form_id", formID,
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"content_type", c.Request().Header.Get("Content-Type"),
		"csrf_token_present", c.Request().Header.Get("X-Csrf-Token") != "",
		"user_agent", c.Request().UserAgent())
}

// processSubmissionRequest processes and validates the submission request
func (h *FormAPIHandler) processSubmissionRequest(c echo.Context, formID string) (model.JSON, error) {
	submissionData, err := h.RequestProcessor.ProcessSubmissionRequest(c)
	if err != nil {
		h.Logger.Error("Failed to process submission request", "form_id", formID, "error", err)

		return nil, h.wrapError("handle submission error", h.ErrorHandler.HandleSubmissionError(c, err))
	}

	h.Logger.Debug("Submission data processed successfully", "form_id", formID, "data_keys", len(submissionData))

	return submissionData, nil
}

// validateSubmissionData validates submission data against form schema
func (h *FormAPIHandler) validateSubmissionData(c echo.Context, form *model.Form, submissionData model.JSON) error {
	validationResult := h.ComprehensiveValidator.ValidateForm(form.Schema, submissionData)
	if !validationResult.IsValid {
		h.Logger.Warn("Form validation failed", "form_id", form.ID, "error_count", len(validationResult.Errors))

		return h.wrapError("build multiple error response",
			h.ResponseBuilder.BuildMultipleErrorResponse(c, validationResult.Errors))
	}

	h.Logger.Debug("Form validation passed", "form_id", form.ID)

	return nil
}

// createAndSubmitForm creates and submits the form
func (h *FormAPIHandler) createAndSubmitForm(
	c echo.Context,
	form *model.Form,
	submissionData model.JSON,
) (*model.FormSubmission, error) {
	submission := &model.FormSubmission{
		FormID:      form.ID,
		Data:        submissionData,
		SubmittedAt: time.Now(),
		Status:      model.SubmissionStatusPending,
	}

	err := h.FormService.SubmitForm(c.Request().Context(), submission)
	if err != nil {
		h.Logger.Error("Failed to submit form", "form_id", form.ID, "submission_id", submission.ID, "error", err)

		return nil, h.wrapError("handle submission error", h.ErrorHandler.HandleSubmissionError(c, err))
	}

	return submission, nil
}

// wrapError provides consistent error wrapping
func (h *FormAPIHandler) wrapError(ctx string, err error) error {
	return fmt.Errorf("%s: %w", ctx, err)
}
