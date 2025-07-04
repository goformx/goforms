package forms

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/infrastructure/adapters/http"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
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

// NewForm handles GET /forms/new
func (h *FormHandler) NewForm(ctx httpiface.Context) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// For now, return a simple success response since GetNewFormPage doesn't exist
	// TODO: Add GetNewFormPage method to FormUseCaseService
	return h.responseAdapter.BuildSuccessResponse(adapterCtx, "New form page loaded", nil)
}

// CreateForm handles POST /forms
func (h *FormHandler) CreateForm(ctx httpiface.Context) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Parse request using adapter
	createReq, err := h.requestAdapter.ParseCreateFormRequest(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse create form request", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid request format"))
	}

	// Call application service
	createResp, err := h.formService.CreateForm(echoCtx.Request().Context(), createReq)
	if err != nil {
		h.logger.Error("failed to create form", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("failed to create form, please try again"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildFormResponse(adapterCtx, createResp)
}

// EditForm handles GET /forms/:id/edit
func (h *FormHandler) EditForm(ctx httpiface.Context) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Parse form ID
	formID, err := h.requestAdapter.ParseFormID(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid form ID"))
	}

	// Call application service
	editResp, err := h.formService.GetForm(echoCtx.Request().Context(), formID)
	if err != nil {
		h.logger.Error("failed to get form for edit", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Failed to load form for editing"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildFormResponse(adapterCtx, editResp)
}

// UpdateForm handles PUT /forms/:id
func (h *FormHandler) UpdateForm(ctx httpiface.Context) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Parse form ID
	formID, err := h.requestAdapter.ParseFormID(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid form ID"))
	}

	// Parse update request
	updateReq, err := h.requestAdapter.ParseUpdateFormRequest(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse update form request", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid request format"))
	}

	// Set form ID in request
	updateReq.ID = formID

	// Call application service
	updateResp, err := h.formService.UpdateForm(echoCtx.Request().Context(), updateReq)
	if err != nil {
		h.logger.Error("failed to update form", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Failed to update form. Please try again."))
	}

	// Build response using adapter
	return h.responseAdapter.BuildFormResponse(adapterCtx, updateResp)
}

// DeleteForm handles DELETE /forms/:id
func (h *FormHandler) DeleteForm(ctx httpiface.Context) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Parse form ID
	formID, err := h.requestAdapter.ParseFormID(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid form ID"))
	}

	// Parse delete request
	deleteReq, err := h.requestAdapter.ParseDeleteFormRequest(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse delete form request", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid request format"))
	}

	// Set form ID in request
	deleteReq.ID = formID

	// Call application service
	deleteResp, err := h.formService.DeleteForm(echoCtx.Request().Context(), deleteReq)
	if err != nil {
		h.logger.Error("failed to delete form", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Failed to delete form. Please try again."))
	}

	// Build response using adapter
	return h.responseAdapter.BuildSuccessResponse(adapterCtx, deleteResp.Message, nil)
}

// FormSubmissions handles GET /forms/:id/submissions
func (h *FormHandler) FormSubmissions(ctx httpiface.Context) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// For now, return a simple success response since GetFormSubmissions doesn't exist
	// TODO: Add GetFormSubmissions method to FormUseCaseService
	return h.responseAdapter.BuildSuccessResponse(adapterCtx, "Form submissions loaded", nil)
}
