package pages

import (
	"fmt"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// PageHandler handles public page routes (e.g., home page)
type PageHandler struct {
	handlers.BaseHandler
	renderer     view.Renderer
	config       *config.Config
	assetManager web.AssetManagerInterface
	logger       logging.Logger
}

// NewPageHandler creates a new PageHandler with a single home route.
func NewPageHandler(
	renderer view.Renderer,
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	logger logging.Logger,
) *PageHandler {
	h := &PageHandler{
		BaseHandler:  *handlers.NewBaseHandler("pages"),
		renderer:     renderer,
		config:       cfg,
		assetManager: assetManager,
		logger:       logger,
	}

	// Register the home page route
	h.AddRoute(httpiface.Route{
		Method:      "GET",
		Path:        "/",
		Handler:     h.handleHome,
		Name:        "home",
		Description: "Home page",
	})

	return h
}

// handleHome is the handler for GET /
func (h *PageHandler) handleHome(ctx httpiface.Context) error {
	// Extract the underlying Echo context for rendering
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return echo.NewHTTPError(500, "Internal server error")
	}

	// Create page data for the home template
	pageData := view.NewPageData(h.config, h.assetManager, echoCtx, "GoFormX - Self-Hosted Form Backend")

	// Render the home page using the generated template
	homeComponent := pages.Home(*pageData)

	if err := h.renderer.Render(echoCtx, homeComponent); err != nil {
		return fmt.Errorf("failed to render home page: %w", err)
	}

	return nil
}
