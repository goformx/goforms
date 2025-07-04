package validation

import (
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
)

// ValidationHandler handles validation schema requests
type ValidationHandler struct {
	handlers.BaseHandler
}

// NewValidationHandler creates a new ValidationHandler and registers all validation routes
func NewValidationHandler() *ValidationHandler {
	h := &ValidationHandler{
		BaseHandler: *handlers.NewBaseHandler("validation"),
	}

	// Form validation schemas
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/validation/forms/new",
		Handler: h.handleNewFormValidation,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/validation/forms/:id",
		Handler: h.handleFormValidation,
	})

	// Auth validation schemas
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/validation/login",
		Handler: h.handleLoginValidation,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/validation/signup",
		Handler: h.handleSignupValidation,
	})

	return h
}

// handleNewFormValidation returns validation schema for new form creation
func (h *ValidationHandler) handleNewFormValidation(ctx httpiface.Context) error {
	// TODO: Implement actual validation schema generation
	schema := map[string]interface{}{
		"title": map[string]interface{}{
			"type":    "required",
			"message": "Form title is required",
		},
	}

	return ctx.JSON(200, schema)
}

// handleFormValidation returns validation schema for a specific form
func (h *ValidationHandler) handleFormValidation(ctx httpiface.Context) error {
	formID := ctx.Param("id")
	if formID == "" {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Form ID is required",
		})
	}

	// TODO: Implement actual form validation schema generation
	// For now, return placeholder response
	clientValidation := map[string]interface{}{
		"form_id": formID,
		"fields":  []interface{}{},
		"rules":   map[string]interface{}{},
	}

	return ctx.JSON(200, clientValidation)
}

// handleLoginValidation returns login form validation schema
func (h *ValidationHandler) handleLoginValidation(ctx httpiface.Context) error {
	// TODO: Implement actual login validation schema
	schema := map[string]interface{}{
		"email": map[string]interface{}{
			"type":    "required|email",
			"message": "Valid email is required",
		},
		"password": map[string]interface{}{
			"type":    "required|min:6",
			"message": "Password must be at least 6 characters",
		},
	}

	return ctx.JSON(200, schema)
}

// handleSignupValidation returns signup form validation schema
func (h *ValidationHandler) handleSignupValidation(ctx httpiface.Context) error {
	// TODO: Implement actual signup validation schema
	schema := map[string]interface{}{
		"name": map[string]interface{}{
			"type":    "required|min:2",
			"message": "Name must be at least 2 characters",
		},
		"email": map[string]interface{}{
			"type":    "required|email|unique:users",
			"message": "Valid and unique email is required",
		},
		"password": map[string]interface{}{
			"type":    "required|min:8|confirmed",
			"message": "Password must be at least 8 characters and confirmed",
		},
		"password_confirmation": map[string]interface{}{
			"type":    "required",
			"message": "Password confirmation is required",
		},
	}

	return ctx.JSON(200, schema)
}
