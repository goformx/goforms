package dashboard

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/dto"
	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/infrastructure/adapters/http"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/view"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
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
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Parse user ID from context (session)
	userID, err := h.requestAdapter.ParseUserID(infraCtx)
	if err != nil {
		h.logger.Warn("authentication required for dashboard access", "error", err)

		return h.responseAdapter.BuildUnauthorizedResponse(infraCtx)
	}

	// Parse pagination request
	paginationReq, err := h.requestAdapter.ParsePaginationRequest(infraCtx)
	if err != nil {
		// Use default pagination if not provided
		paginationReq = &dto.PaginationRequest{
			Page:  1,
			Limit: 10,
		}
	}

	// Call application service
	dashboardResp, err := h.formService.ListForms(ctx.RequestContext(), userID, paginationReq)
	if err != nil {
		h.logger.Error("failed to fetch user forms", "user_id", userID, "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to load your forms, please try again"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildFormListResponse(infraCtx, dashboardResp)
}
