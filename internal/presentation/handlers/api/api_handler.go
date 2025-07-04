package api

import (
	"fmt"

	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
)

// ApiHandler handles all API form-related routes
// Implements httpiface.Handler
type APIHandler struct {
	handlers.BaseHandler
}

// NewApiHandler creates a new ApiHandler and registers all API routes
func NewAPIHandler() *APIHandler {
	h := &APIHandler{
		BaseHandler: *handlers.NewBaseHandler("api"),
	}

	// Authenticated API routes
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/forms",
		Handler: h.ListForms,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/forms/:id",
		Handler: h.GetForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/forms/:id/schema",
		Handler: h.GetFormSchema,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/forms/:id/validation",
		Handler: h.GetFormValidationSchema,
	})
	h.AddRoute(httpiface.Route{
		Method:  "PUT",
		Path:    "/api/v1/forms/:id/schema",
		Handler: h.UpdateFormSchema,
	})
	h.AddRoute(httpiface.Route{
		Method:  "POST",
		Path:    "/api/v1/forms/:id/submit",
		Handler: h.SubmitForm,
	})

	return h
}

// ListForms handles GET /api/v1/forms
func (h *APIHandler) ListForms(ctx httpiface.Context) error {
	return fmt.Errorf("list forms (placeholder)")
}

// GetForm handles GET /api/v1/forms/:id
func (h *APIHandler) GetForm(ctx httpiface.Context) error {
	return fmt.Errorf("get form (placeholder)")
}

// GetFormSchema handles GET /api/v1/forms/:id/schema
func (h *APIHandler) GetFormSchema(ctx httpiface.Context) error {
	return fmt.Errorf("get form schema (placeholder)")
}

// GetFormValidationSchema handles GET /api/v1/forms/:id/validation
func (h *APIHandler) GetFormValidationSchema(ctx httpiface.Context) error {
	return fmt.Errorf("get form validation schema (placeholder)")
}

// UpdateFormSchema handles PUT /api/v1/forms/:id/schema
func (h *APIHandler) UpdateFormSchema(ctx httpiface.Context) error {
	return fmt.Errorf("update form schema (placeholder)")
}

// SubmitForm handles POST /api/v1/forms/:id/submit
func (h *APIHandler) SubmitForm(ctx httpiface.Context) error {
	return fmt.Errorf("submit form (placeholder)")
}
