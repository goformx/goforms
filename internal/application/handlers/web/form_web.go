// internal/application/handlers/web/form_web.go
package web

import (
	"context"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/labstack/echo/v4"
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
func (h *FormWebHandler) RegisterRoutes(e *echo.Echo, accessManager *access.AccessManager) {
	forms := e.Group(constants.PathForms)
	forms.Use(access.Middleware(accessManager, h.Logger))

	forms.GET("/new", h.handleNew)
	forms.POST("", h.handleCreate)
	forms.GET("/:id/edit", h.handleEdit)
	forms.POST("/:id/edit", h.handleUpdate)
	forms.DELETE("/:id", h.handleDelete)
	forms.GET("/:id/submissions", h.handleSubmissions)
}

// Register satisfies the Handler interface
func (h *FormWebHandler) Register(e *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// Start satisfies the Handler interface
func (h *FormWebHandler) Start(ctx context.Context) error {
	return nil // No initialization needed
}

// Stop satisfies the Handler interface
func (h *FormWebHandler) Stop(ctx context.Context) error {
	return nil // No cleanup needed
}
