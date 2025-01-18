package web

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/components"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/view"
)

type PageHandler struct {
	renderer *view.Renderer
	logger   logger.Logger
}

func NewPageHandler(renderer *view.Renderer, logger logger.Logger) *PageHandler {
	return &PageHandler{renderer, logger}
}

func (h *PageHandler) wrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

func (h *PageHandler) Home(c echo.Context) error {
	if err := h.renderer.Render(c, components.Home()); err != nil {
		h.logger.Error("failed to render home page", logger.Error(err))
		return h.wrapError(err, "failed to render home page")
	}
	return nil
}

func (h *PageHandler) Contact(c echo.Context) error {
	if err := h.renderer.Render(c, components.Contact()); err != nil {
		h.logger.Error("failed to render contact page", logger.Error(err))
		return h.wrapError(err, "failed to render contact page")
	}
	return nil
}

func (h *PageHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/", h.Home)
	e.GET("/contact", h.Contact)
}
