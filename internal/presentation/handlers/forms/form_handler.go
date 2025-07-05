package forms

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/dto"
	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/adapters/http"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/view"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
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

//nolint:funlen // its fine, move along
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
		Method:  "GET",
		Path:    "/forms/:id/preview",
		Handler: h.PreviewForm,
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
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/forms/:id/schema",
		Handler: h.GetFormSchema,
	})
	h.AddRoute(httpiface.Route{
		Method:  "PUT",
		Path:    "/api/v1/forms/:id/schema",
		Handler: h.UpdateFormSchema,
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

// getFormAndContext is a helper function to get form data and context for rendering
func (h *FormHandler) getFormAndContext(
	ctx httpiface.Context,
	_ string, // formID is not used since we parse it from the context
) (*model.Form, *http.EchoContextAdapter, error) {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return nil, nil, fmt.Errorf("internal server error: context conversion failed")
	}

	// Parse form ID
	parsedFormID, err := h.requestAdapter.ParseFormID(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)

		return nil, nil, fmt.Errorf("invalid form ID")
	}

	// Call application service
	formResp, err := h.formService.GetForm(ctx.RequestContext(), parsedFormID)
	if err != nil {
		h.logger.Error("failed to get form", "form_id", parsedFormID, "error", err)

		return nil, nil, fmt.Errorf("failed to load form")
	}

	// Convert DTO to domain model for template rendering
	form := &model.Form{
		ID:          formResp.ID,
		Title:       formResp.Title,
		Description: formResp.Description,
		Schema:      formResp.Schema,
		UserID:      formResp.UserID,
		Status:      formResp.Status,
		CreatedAt:   formResp.CreatedAt,
		UpdatedAt:   formResp.UpdatedAt,
	}

	// Get the underlying Echo context for rendering
	echoCtx, ok := infraCtx.(*http.EchoContextAdapter)
	if !ok {
		return nil, nil, fmt.Errorf("invalid context type for rendering")
	}

	return form, echoCtx, nil
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
	if buildErr := h.responseAdapter.BuildSuccessResponse(infraCtx, "New form page loaded", nil); buildErr != nil {
		return fmt.Errorf("failed to build success response: %w", buildErr)
	}

	return nil
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

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid request format"))
	}
	// Call application service
	createResp, err := h.formService.CreateForm(ctx.RequestContext(), createReq)
	if err != nil {
		h.logger.Error("failed to create form", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to create form, please try again"))
	}
	// Build response using adapter
	if buildErr := h.responseAdapter.BuildFormResponse(infraCtx, createResp); buildErr != nil {
		return fmt.Errorf("failed to build form response: %w", buildErr)
	}

	return nil
}

// EditForm handles GET /forms/:id/edit
func (h *FormHandler) EditForm(ctx httpiface.Context) error {
	form, echoCtx, err := h.getFormAndContext(ctx, "")
	if err != nil {
		// Get infrastructure context for error response
		infraCtx, getCtxErr := h.getInfraContext(ctx)
		if getCtxErr == nil && infraCtx != nil {
			return h.responseAdapter.BuildErrorResponse(infraCtx, err)
		}

		return err
	}

	// Create page data for template rendering
	pageData := view.NewPageData(h.config, h.assetManager, echoCtx.Context, "Edit Form - "+form.Title)
	pageData.Description = "Edit your form settings and fields"
	pageData.Form = form
	pageData.FormBuilderAssetPath = h.assetManager.AssetPath("src/js/pages/form-builder.ts")

	// Render the edit form template
	if renderErr := h.renderer.Render(echoCtx.Context, pages.EditForm(*pageData, form)); renderErr != nil {
		return fmt.Errorf("failed to render edit form template: %w", renderErr)
	}

	return nil
}

// PreviewForm handles GET /forms/:id/preview
func (h *FormHandler) PreviewForm(ctx httpiface.Context) error {
	form, echoCtx, err := h.getFormAndContext(ctx, "")
	if err != nil {
		// Get infrastructure context for error response
		infraCtx, getCtxErr := h.getInfraContext(ctx)
		if getCtxErr == nil && infraCtx != nil {
			return h.responseAdapter.BuildErrorResponse(infraCtx, err)
		}

		return err
	}

	// Create page data for template rendering
	pageData := view.NewPageData(h.config, h.assetManager, echoCtx.Context, "Form Preview - "+form.Title)
	pageData.Description = "Preview your form as users will see it"
	pageData.Form = form
	pageData.FormPreviewAssetPath = h.assetManager.AssetPath("src/js/pages/form-preview.ts")

	// Render the form preview template
	if renderErr := h.renderer.Render(echoCtx.Context, pages.FormPreview(*pageData, form)); renderErr != nil {
		return fmt.Errorf("failed to render form preview template: %w", renderErr)
	}

	return nil
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

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid form ID"))
	}
	// Parse update request
	updateReq, err := h.requestAdapter.ParseUpdateFormRequest(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse update form request", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid request format"))
	}
	// Set form ID in request
	updateReq.ID = formID
	// Call application service
	updateResp, err := h.formService.UpdateForm(ctx.RequestContext(), updateReq)
	if err != nil {
		h.logger.Error("failed to update form", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to update form, please try again"))
	}
	// Build response using adapter
	if buildErr := h.responseAdapter.BuildFormResponse(infraCtx, updateResp); buildErr != nil {
		return fmt.Errorf("failed to build form response: %w", buildErr)
	}

	return nil
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

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid form ID"))
	}
	// Parse delete request
	deleteReq, err := h.requestAdapter.ParseDeleteFormRequest(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse delete form request", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid request format"))
	}
	// Set form ID in request
	deleteReq.ID = formID
	// Call application service
	deleteResp, err := h.formService.DeleteForm(ctx.RequestContext(), deleteReq)
	if err != nil {
		h.logger.Error("failed to delete form", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to delete form, please try again"))
	}
	// Build response using adapter
	if buildErr := h.responseAdapter.BuildSuccessResponse(infraCtx, deleteResp.Message, nil); buildErr != nil {
		return fmt.Errorf("failed to build success response: %w", buildErr)
	}

	return nil
}

// FormSubmissions handles GET /forms/:id/submissions
func (h *FormHandler) FormSubmissions(ctx httpiface.Context) error {
	form, echoCtx, err := h.getFormAndContext(ctx, "")
	if err != nil {
		// Get infrastructure context for error response
		infraCtx, getCtxErr := h.getInfraContext(ctx)
		if getCtxErr == nil && infraCtx != nil {
			return h.responseAdapter.BuildErrorResponse(infraCtx, err)
		}

		return err
	}

	// Create page data for template rendering
	pageData := view.NewPageData(h.config, h.assetManager, echoCtx.Context, "Form Submissions - "+form.Title)
	pageData.Description = "View and manage form submissions"
	pageData.Form = form
	// For now, we'll show an empty submissions list since the service doesn't exist yet
	pageData.Submissions = []*model.FormSubmission{}

	// Render the form submissions template
	if renderErr := h.renderer.Render(echoCtx.Context, pages.FormSubmissions(*pageData)); renderErr != nil {
		return fmt.Errorf("failed to render form submissions template: %w", renderErr)
	}

	return nil
}

// GetFormSchema handles GET /api/v1/forms/:id/schema
func (h *FormHandler) GetFormSchema(ctx httpiface.Context) error {
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

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid form ID"))
	}
	// Call application service
	editResp, err := h.formService.GetForm(ctx.RequestContext(), formID)
	if err != nil {
		h.logger.Error("failed to get form schema", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to load form schema"))
	}
	// Return just the schema as JSON
	return infraCtx.JSON(200, editResp.Schema)
}

// UpdateFormSchema handles PUT /api/v1/forms/:id/schema
func (h *FormHandler) UpdateFormSchema(ctx httpiface.Context) error {
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

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid form ID"))
	}

	// Get the existing form first to preserve its data
	existingForm, err := h.formService.GetForm(ctx.RequestContext(), formID)
	if err != nil {
		h.logger.Error("failed to get existing form", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to get form"))
	}

	// Parse the schema from request body
	var schema map[string]any

	echoCtx, ok := infraCtx.(*http.EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type for binding")
	}

	if bindErr := echoCtx.Bind(&schema); bindErr != nil {
		h.logger.Error("failed to parse schema", "error", bindErr)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid schema format"))
	}

	// Create update request preserving existing form data
	updateReq := &dto.UpdateFormRequest{
		ID:          formID,
		Title:       existingForm.Title,
		Description: existingForm.Description,
		Schema:      schema,
		Status:      existingForm.Status,
	}

	// Get user ID from context
	userID := infraCtx.Get("user_id")
	if userID != nil {
		if userIDStr, userIDOk := userID.(string); userIDOk {
			updateReq.UserID = userIDStr
		}
	}

	// Call application service
	updateResp, err := h.formService.UpdateForm(ctx.RequestContext(), updateReq)
	if err != nil {
		h.logger.Error("failed to update form schema", "form_id", formID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to update form schema"))
	}

	// Return the updated schema
	return infraCtx.JSON(200, updateResp.Schema)
}
