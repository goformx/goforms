package pages

import (
	"fmt"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/view"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
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
	// Create page data for the home template
	pageData := &view.PageData{
		Title:       "GoFormX - Self-Hosted Form Backend",
		Description: "Welcome to GoFormX!",
		Version:     h.config.App.Version,
		Environment: h.config.App.Environment,
		Config:      h.config,
		AssetPath:   h.assetManager.AssetPath,
	}

	// Render the home page using the framework-agnostic interface
	homeComponent := pages.Home(*pageData)

	if renderErr := ctx.RenderComponent(homeComponent); renderErr != nil {
		return fmt.Errorf("failed to render home component: %w", renderErr)
	}

	return nil
}
