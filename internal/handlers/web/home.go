package web

import (
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/jonesrussell/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// HomeHandler handles the homepage routes
type HomeHandler struct {
	base handlers.Base
	Renderer *view.Renderer
}

// NewHomeHandler creates a new HomeHandler
func NewHomeHandler(logger logging.Logger, renderer *view.Renderer) *HomeHandler {
	return &HomeHandler{
		base: handlers.Base{
			Logger: logger,
		},
		Renderer: renderer,
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
		Title: "GoForms - Free Form Backend Service",
	}

	return h.Renderer.Render(c, pages.Home(data))
} 