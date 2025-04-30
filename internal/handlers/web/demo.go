package web

import (
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/jonesrussell/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// DemoHandler handles the demo page routes
type DemoHandler struct {
	base handlers.Base
	Renderer *view.Renderer
	SubscriptionService subscription.Service
}

// NewDemoHandler creates a new DemoHandler
func NewDemoHandler(
	logger logging.Logger,
	renderer *view.Renderer,
	subscriptionService subscription.Service,
) *DemoHandler {
	return &DemoHandler{
		base: handlers.Base{
			Logger: logger,
		},
		Renderer: renderer,
		SubscriptionService: subscriptionService,
	}
}

// Register sets up the routes for the demo handler
func (h *DemoHandler) Register(e *echo.Echo) {
	h.base.RegisterRoute(e, "GET", "/demo", h.handleDemo)
}

// handleDemo renders the demo page
func (h *DemoHandler) handleDemo(c echo.Context) error {
	h.base.Logger.Debug("handling demo page request")

	data := shared.PageData{
		Title: "GoForms Demo - See it in Action",
	}

	return h.Renderer.Render(c, pages.Demo(data))
} 