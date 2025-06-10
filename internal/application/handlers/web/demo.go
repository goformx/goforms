package web

import (
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// DemoHandler handles demo page requests
type DemoHandler struct {
	HandlerDeps
}

// NewDemoHandler creates a new demo handler using HandlerDeps
func NewDemoHandler(deps HandlerDeps) (*DemoHandler, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}
	return &DemoHandler{HandlerDeps: deps}, nil
}

// Register registers the demo routes
func (h *DemoHandler) Register(e *echo.Echo) {
	e.GET("/demo", h.handleDemo)
}

// handleDemo handles the demo page request
func (h *DemoHandler) handleDemo(c echo.Context) error {
	data := shared.BuildPageData(h.Config, "Demo")
	return h.Renderer.Render(c, pages.Demo(data))
}
