package web

import (
	"github.com/goformx/goforms/internal/handlers"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// HomeHandler handles the homepage routes
type HomeHandler struct {
	base     handlers.Base
	logger   logging.Logger
	renderer *view.Renderer
}

// NewHomeHandler creates a new HomeHandler
func NewHomeHandler(logger logging.Logger, renderer *view.Renderer) *HomeHandler {
	return &HomeHandler{
		base: handlers.Base{
			Logger: logger,
		},
		logger:   logger,
		renderer: renderer,
	}
}

// Register sets up the routes for the home handler
func (h *HomeHandler) Register(e *echo.Echo) {
	h.base.RegisterRoute(e, "GET", "/", h.handleHome)
}

// handleHome renders the home page
func (h *HomeHandler) handleHome(c echo.Context) error {
	h.base.Logger.Debug("handling home page request")

	data := shared.PageData{
		Title:     "GoFormX - Free Form Backend Service",
		CSRFToken: "", // No CSRF token needed for homepage
	}

	return h.renderer.Render(c, pages.Home(data))
}
