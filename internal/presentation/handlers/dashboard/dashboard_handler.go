package dashboard

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

// DashboardHandler handles the /dashboard route
// Implements httpiface.Handler
type DashboardHandler struct {
	handlers.BaseHandler
	formService     *services.FormUseCaseService
	requestAdapter  http.RequestAdapter
	responseAdapter http.ResponseAdapter
	renderer        view.Renderer
	config          *config.Config
	assetManager    web.AssetManagerInterface
	logger          logging.Logger
}

// NewDashboardHandler creates a new DashboardHandler and registers the /dashboard route
func NewDashboardHandler(
	formService *services.FormUseCaseService,
	requestAdapter http.RequestAdapter,
	responseAdapter http.ResponseAdapter,
	renderer view.Renderer,
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	logger logging.Logger,
) *DashboardHandler {
	h := &DashboardHandler{
		BaseHandler:     *handlers.NewBaseHandler("dashboard"),
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
		Path:    "/dashboard",
		Handler: h.Dashboard,
	})

	return h
}

// getInfraContext is a simple bridge to convert presentation Context to infrastructure Context
func (h *DashboardHandler) getInfraContext(ctx httpiface.Context) (http.Context, error) {
	// Simple type assertion to the infrastructure adapter
	if infraCtx, ok := ctx.(*http.EchoContextAdapter); ok {
		return infraCtx, nil
	}

	return nil, fmt.Errorf("invalid context type")
}

// Dashboard handles GET /dashboard
func (h *DashboardHandler) Dashboard(ctx httpiface.Context) error {
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}

	userID, err := h.requestAdapter.ParseUserID(infraCtx)
	if err != nil {
		h.logger.Warn("authentication required for dashboard access", "error", err)

		if unauthorizedErr := h.responseAdapter.BuildUnauthorizedResponse(infraCtx); unauthorizedErr != nil {
			return fmt.Errorf("failed to build unauthorized response: %w", unauthorizedErr)
		}

		return nil
	}

	paginationReq, err := h.requestAdapter.ParsePaginationRequest(infraCtx)
	if err != nil {
		paginationReq = &dto.PaginationRequest{Page: 1, Limit: 10}
	}

	dashboardResp, err := h.formService.ListForms(ctx.RequestContext(), userID, paginationReq)
	if err != nil {
		h.logger.Error("failed to fetch user forms", "user_id", userID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to load your forms, please try again"))
	}

	return h.renderDashboard(infraCtx, dashboardResp)
}

func (h *DashboardHandler) renderDashboard(
	infraCtx http.Context,
	dashboardResp *dto.FormListResponse,
) error {
	echoCtx, ok := infraCtx.(*http.EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type for rendering")
	}

	pageData := view.NewPageData(h.config, h.assetManager, echoCtx.Context, "Dashboard")

	forms := make([]*model.Form, len(dashboardResp.Forms))
	for i, formDTO := range dashboardResp.Forms {
		forms[i] = &model.Form{
			ID:          formDTO.ID,
			Title:       formDTO.Title,
			Description: formDTO.Description,
			Schema:      formDTO.Schema,
			UserID:      formDTO.UserID,
			Status:      formDTO.Status,
			CreatedAt:   formDTO.CreatedAt,
			UpdatedAt:   formDTO.UpdatedAt,
		}
	}

	if err := h.renderer.Render(echoCtx.Context, pages.Dashboard(*pageData, forms)); err != nil {
		h.logger.Error("failed to render dashboard template", "error", err)

		return fmt.Errorf("failed to render dashboard: %w", err)
	}

	return nil
}
