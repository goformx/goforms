package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/pages"
)

type PageHandler struct {
	logger logger.Logger
}

func NewPageHandler(logger logger.Logger) *PageHandler {
	return &PageHandler{
		logger: logger,
	}
}

func (h *PageHandler) Home(c echo.Context) error {
	component := pages.Home()
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *PageHandler) Contact(c echo.Context) error {
	component := pages.Contact()
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *PageHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/", h.Home)
	e.GET("/contact", h.Contact)
}
