package web

import (
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type DemoHandler struct {
	HandlerDeps
}

func NewDemoHandler(deps HandlerDeps) *DemoHandler {
	return &DemoHandler{HandlerDeps: deps}
}

func (h *DemoHandler) Register(e *echo.Echo) {
	e.GET("/demo", h.handleDemo)
}

func (h *DemoHandler) handleDemo(c echo.Context) error {
	data := shared.BuildPageData(h.Config, "Demo")
	return h.Renderer.Render(c, pages.Demo(data))
}
