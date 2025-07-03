package web

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
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

// handleListForms and handleGetForm moved to form_api_authenticated.go

// handleFormSchema moved to form_api_public.go

// handleFormSchemaUpdate moved to form_api_authenticated.go

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

// Helper methods moved to form_submission_handler.go
