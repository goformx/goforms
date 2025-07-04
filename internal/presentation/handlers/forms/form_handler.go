package forms

import (
	"fmt"

	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
)

// FormHandler handles all form-related routes
// Implements httpiface.Handler
type FormHandler struct {
	handlers.BaseHandler
}

// NewFormHandler creates a new FormHandler and registers all form routes
func NewFormHandler() *FormHandler {
	h := &FormHandler{
		BaseHandler: *handlers.NewBaseHandler("forms"),
	}

	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/forms/new",
		Handler: h.NewForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "POST",
		Path:    "/forms",
		Handler: h.CreateForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/forms/:id/edit",
		Handler: h.EditForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "PUT",
		Path:    "/forms/:id",
		Handler: h.UpdateForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "DELETE",
		Path:    "/forms/:id",
		Handler: h.DeleteForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/forms/:id/submissions",
		Handler: h.FormSubmissions,
	})

	return h
}

// NewForm handles GET /forms/new
func (h *FormHandler) NewForm(ctx httpiface.Context) error {
	return fmt.Errorf("New form page (placeholder)")
}

// CreateForm handles POST /forms
func (h *FormHandler) CreateForm(ctx httpiface.Context) error {
	return fmt.Errorf("Create form (placeholder)")
}

// EditForm handles GET /forms/:id/edit
func (h *FormHandler) EditForm(ctx httpiface.Context) error {
	return fmt.Errorf("Edit form page (placeholder)")
}

// UpdateForm handles PUT /forms/:id
func (h *FormHandler) UpdateForm(ctx httpiface.Context) error {
	return fmt.Errorf("Update form (placeholder)")
}

// DeleteForm handles DELETE /forms/:id
func (h *FormHandler) DeleteForm(ctx httpiface.Context) error {
	return fmt.Errorf("Delete form (placeholder)")
}

// FormSubmissions handles GET /forms/:id/submissions
func (h *FormHandler) FormSubmissions(ctx httpiface.Context) error {
	return fmt.Errorf("Form submissions page (placeholder)")
}
