package forms

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/infrastructure/adapters/http"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/view"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
)

// FormHandler handles all form-related routes
// Implements httpiface.Handler
type FormHandler struct {
	handlers.BaseHandler
	formService     *services.FormUseCaseService
	requestAdapter  http.RequestAdapter
	responseAdapter http.ResponseAdapter
	renderer        view.Renderer
	config          *config.Config
	assetManager    web.AssetManagerInterface
	logger          logging.Logger
}

// NewFormHandler creates a new FormHandler and registers all form routes
func NewFormHandler(
	formService *services.FormUseCaseService,
	requestAdapter http.RequestAdapter,
	responseAdapter http.ResponseAdapter,
	renderer view.Renderer,
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	logger logging.Logger,
) *FormHandler {
	h := &FormHandler{
		BaseHandler:     *handlers.NewBaseHandler("forms"),
		formService:     formService,
		requestAdapter:  requestAdapter,
		responseAdapter: responseAdapter,
		renderer:        renderer,
		config:          cfg,
		assetManager:    assetManager,
		logger:          logger,
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

// getInfraContext is a simple bridge to convert presentation Context to infrastructure Context
func (h *FormHandler) getInfraContext(ctx httpiface.Context) (http.Context, error) {
	// Simple type assertion to the infrastructure adapter
	if infraCtx, ok := ctx.(*http.EchoContextAdapter); ok {
		return infraCtx, nil
	}

	return nil, fmt.Errorf("invalid context type")
}

// NewForm handles GET /forms/new
func (h *FormHandler) NewForm(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}
	// For now, return a simple success response since GetNewFormPage doesn't exist
	// TODO: Add GetNewFormPage method to FormUseCaseService
	return h.responseAdapter.BuildSuccessResponse(infraCtx, "New form page loaded", nil)
}

// CreateForm handles POST /forms
func (h *FormHandler) CreateForm(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}
	// Parse request using adapter
	createReq, err := h.requestAdapter.ParseCreateFormRequest(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse create form request", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Invalid request format"))
	}
	// Call application service
	createResp, err := h.formService.CreateForm(ctx.RequestContext(), createReq)
	if err != nil {
		h.logger.Error("failed to create form", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to create form, please try again"))
	}
	// Build response using adapter
	return h.responseAdapter.BuildFormResponse(infraCtx, createResp)
}

// EditForm handles GET /forms/:id/edit
func (h *FormHandler) EditForm(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}
	// Parse form ID
	formID, err := h.requestAdapter.ParseFormID(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Invalid form ID"))
	}
	// Call application service
	editResp, err := h.formService.GetForm(ctx.RequestContext(), formID)
	if err != nil {
		h.logger.Error("failed to get form for edit", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Failed to load form for editing"))
	}
	// Build response using adapter
	return h.responseAdapter.BuildFormResponse(infraCtx, editResp)
}

// UpdateForm handles PUT /forms/:id
func (h *FormHandler) UpdateForm(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}
	// Parse form ID
	formID, err := h.requestAdapter.ParseFormID(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Invalid form ID"))
	}
	// Parse update request
	updateReq, err := h.requestAdapter.ParseUpdateFormRequest(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse update form request", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Invalid request format"))
	}
	// Set form ID in request
	updateReq.ID = formID
	// Call application service
	updateResp, err := h.formService.UpdateForm(ctx.RequestContext(), updateReq)
	if err != nil {
		h.logger.Error("failed to update form", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Failed to update form. Please try again."))
	}
	// Build response using adapter
	return h.responseAdapter.BuildFormResponse(infraCtx, updateResp)
}

// DeleteForm handles DELETE /forms/:id
func (h *FormHandler) DeleteForm(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}
	// Parse form ID
	formID, err := h.requestAdapter.ParseFormID(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Invalid form ID"))
	}
	// Parse delete request
	deleteReq, err := h.requestAdapter.ParseDeleteFormRequest(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse delete form request", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Invalid request format"))
	}
	// Set form ID in request
	deleteReq.ID = formID
	// Call application service
	deleteResp, err := h.formService.DeleteForm(ctx.RequestContext(), deleteReq)
	if err != nil {
		h.logger.Error("failed to delete form", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("Failed to delete form. Please try again."))
	}
	// Build response using adapter
	return h.responseAdapter.BuildSuccessResponse(infraCtx, deleteResp.Message, nil)
}

// FormSubmissions handles GET /forms/:id/submissions
func (h *FormHandler) FormSubmissions(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}
	// For now, return a simple success response since GetFormSubmissions doesn't exist
	// TODO: Add GetFormSubmissions method to FormUseCaseService
	return h.responseAdapter.BuildSuccessResponse(infraCtx, "Form submissions loaded", nil)
}
