package dashboard

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/dto"
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

// Dashboard handles GET /dashboard
func (h *DashboardHandler) Dashboard(ctx httpiface.Context) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Parse user ID from context (session)
	userID, err := h.requestAdapter.ParseUserID(adapterCtx)
	if err != nil {
		h.logger.Warn("authentication required for dashboard access", "error", err)

		return h.responseAdapter.BuildUnauthorizedResponse(adapterCtx)
	}

	// Parse pagination request
	paginationReq, err := h.requestAdapter.ParsePaginationRequest(adapterCtx)
	if err != nil {
		// Use default pagination if not provided
		paginationReq = &dto.PaginationRequest{
			Page:  1,
			Limit: 10,
		}
	}

	// Call application service
	dashboardResp, err := h.formService.ListForms(echoCtx.Request().Context(), userID, paginationReq)
	if err != nil {
		h.logger.Error("failed to fetch user forms", "user_id", userID, "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("failed to load your forms, please try again"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildFormListResponse(adapterCtx, dashboardResp)
}
