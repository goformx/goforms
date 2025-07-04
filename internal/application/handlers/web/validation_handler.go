package web

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/validation"
)

// ValidationHandler handles validation schema requests
type ValidationHandler struct {
	*BaseHandler
	SchemaGenerator *validation.SchemaGenerator
}

// NewValidationHandler creates a new ValidationHandler
func NewValidationHandler(
	base *BaseHandler,
	schemaGenerator *validation.SchemaGenerator,
) *ValidationHandler {
	return &ValidationHandler{
		BaseHandler:     base,
		SchemaGenerator: schemaGenerator,
	}
}

// RegisterRoutes registers validation routes
func (h *ValidationHandler) RegisterRoutes(e *echo.Echo) {
	// API validation routes
	api := e.Group(constants.PathAPIv1)
	validationGroup := api.Group(constants.PathAPIValidation)

	// Form validation schemas
	validationGroup.GET("/forms/new", h.handleNewFormValidation)
	validationGroup.GET("/forms/:id", h.handleFormValidation)

	// Auth validation schemas
	validationGroup.GET("/login", h.handleLoginValidation)
	validationGroup.GET("/signup", h.handleSignupValidation)
}

// Register satisfies the Handler interface
func (h *ValidationHandler) Register(_ *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// Start initializes the validation handler
func (h *ValidationHandler) Start(_ context.Context) error {
	return nil
}

// Stop cleans up the validation handler
func (h *ValidationHandler) Stop(_ context.Context) error {
	return nil
}

// handleNewFormValidation returns validation schema for new form creation
func (h *ValidationHandler) handleNewFormValidation(c echo.Context) error {
	schema := map[string]any{
		"title": map[string]any{
			"type":    "required",
			"message": "Form title is required",
		},
	}

	return response.Success(c, schema)
}

// handleFormValidation returns validation schema for a specific form
func (h *ValidationHandler) handleFormValidation(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return response.ErrorResponse(c, 400, "Form ID is required")
	}

	// Get form and return its validation schema
	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form for validation", "error", err, "form_id", formID)

		return response.ErrorResponse(c, 404, "Form not found")
	}

	// Generate client-side validation rules from form schema
	comprehensiveValidator := validation.NewComprehensiveValidator()

	clientValidation, err := comprehensiveValidator.GenerateClientValidation(form.Schema)
	if err != nil {
		h.Logger.Error("failed to generate client validation schema", "error", err, "form_id", formID)

		return response.ErrorResponse(c, 500, "Failed to generate validation schema")
	}

	return response.Success(c, clientValidation)
}

// handleLoginValidation returns login form validation schema
func (h *ValidationHandler) handleLoginValidation(c echo.Context) error {
	schema := h.SchemaGenerator.GenerateLoginSchema()

	return response.Success(c, schema)
}

// handleSignupValidation returns signup form validation schema
func (h *ValidationHandler) handleSignupValidation(c echo.Context) error {
	schema := h.SchemaGenerator.GenerateSignupSchema()

	return response.Success(c, schema)
}
