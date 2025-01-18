package handlers

import (
	"github.com/jonesrussell/goforms/internal/components"
	"github.com/labstack/echo/v4"
)

type MarketingHandler struct{}

func NewMarketingHandler() *MarketingHandler {
	return &MarketingHandler{}
}

func (h *MarketingHandler) Register(e *echo.Echo) {
	e.GET("/", h.HandleHome)
	e.GET("/contact", h.HandleContact)
}

func (h *MarketingHandler) HandleHome(c echo.Context) error {
	return components.Home().Render(c.Request().Context(), c.Response().Writer)
}

func (h *MarketingHandler) HandleContact(c echo.Context) error {
	return components.Contact().Render(c.Request().Context(), c.Response().Writer)
}
