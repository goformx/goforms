package web

import (
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// WebHandler handles web page requests
type WebHandler struct {
	HandlerDeps
}

// NewWebHandler creates a new web handler using HandlerDeps
func NewWebHandler(deps HandlerDeps) (*WebHandler, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}
	return &WebHandler{HandlerDeps: deps}, nil
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	e.GET("/", h.handleHome)
}

// handleHome handles the home page request
func (h *WebHandler) handleHome(c echo.Context) error {
	data := shared.BuildPageData(h.Config, c, "Home")
	if err := h.Renderer.Render(c, pages.Home(data)); err != nil {
		data.Message = &shared.Message{
			Type: "error",
			Text: err.Error(),
		}
		return pages.Error(data).Render(c.Request().Context(), c.Response().Writer)
	}
	return nil
}
