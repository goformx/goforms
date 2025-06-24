package web

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/labstack/echo/v4"
)

// FormAPIHandler handles API form operations
type FormAPIHandler struct {
	*FormBaseHandler
	AccessManager    *access.AccessManager
	RequestProcessor FormRequestProcessor
	ResponseBuilder  FormResponseBuilder
	ErrorHandler     FormErrorHandler
}

func NewFormAPIHandler(
	base *BaseHandler,
	formService formdomain.Service,
	accessManager *access.AccessManager,
	formValidator *validation.FormValidator,
	sanitizer sanitization.ServiceInterface,
) *FormAPIHandler {
	// Create dependencies
	requestProcessor := NewFormRequestProcessor(sanitizer, formValidator)
	responseBuilder := NewFormResponseBuilder()
	errorHandler := NewFormErrorHandler(responseBuilder)

	return &FormAPIHandler{
		FormBaseHandler:  NewFormBaseHandler(base, formService, formValidator),
		AccessManager:    accessManager,
		RequestProcessor: requestProcessor,
		ResponseBuilder:  responseBuilder,
		ErrorHandler:     errorHandler,
	}
}

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
	formsAPI.POST("/:id/submit", h.handleFormSubmit)
}

// Register satisfies the Handler interface
func (h *FormAPIHandler) Register(e *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// GET /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchema(c echo.Context) error {
	form, err := h.GetFormByID(c)
	if err != nil {
		return h.HandleError(c, err, "Failed to get form schema")
	}

	return h.ResponseBuilder.BuildSchemaResponse(c, form.Schema)
}

// PUT /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchemaUpdate(c echo.Context) error {
	_, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return h.ErrorHandler.HandleOwnershipError(c, err)
	}

	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return h.HandleError(c, err, "Unauthorized or form not found")
	}

	// Process and validate schema update request
	schema, err := h.RequestProcessor.ProcessSchemaUpdateRequest(c)
	if err != nil {
		return h.ErrorHandler.HandleSchemaError(c, err)
	}

	// Update form schema
	form.Schema = schema
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form); updateErr != nil {
		h.Logger.Error("failed to update form schema", "error", updateErr)
		return h.ErrorHandler.HandleSchemaError(c, updateErr)
	}

	return h.ResponseBuilder.BuildSchemaResponse(c, form.Schema)
}

// POST /api/v1/forms/:id/submit
func (h *FormAPIHandler) handleFormSubmit(c echo.Context) error {
	form, err := h.GetFormByID(c)
	if err != nil {
		return h.HandleError(c, err, "Failed to get form for submission")
	}

	// Process and validate submission request
	submissionData, err := h.RequestProcessor.ProcessSubmissionRequest(c)
	if err != nil {
		return h.ErrorHandler.HandleSubmissionError(c, err)
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
		return h.ErrorHandler.HandleSubmissionError(c, err)
	}

	return h.ResponseBuilder.BuildSubmissionResponse(c, submission)
}

// Start initializes the form API handler.
// This is called during application startup.
func (h *FormAPIHandler) Start(ctx context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the form API handler.
// This is called during application shutdown.
func (h *FormAPIHandler) Stop(ctx context.Context) error {
	return nil // No cleanup needed
}
