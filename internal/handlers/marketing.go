package handlers

import (
	"github.com/jonesrussell/goforms/internal/components"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/labstack/echo/v4"
)

type MarketingHandler struct {
	logger logger.Logger
}

func NewMarketingHandler(logger logger.Logger) *MarketingHandler {
	return &MarketingHandler{
		logger: logger,
	}
}

func (h *MarketingHandler) Register(e *echo.Echo) {
	h.logger.Debug("Registering marketing routes")
	e.GET("/", h.HandleHome)
	e.GET("/contact", h.HandleContact)
}

func (h *MarketingHandler) HandleHome(c echo.Context) error {
	h.logger.Debug("Handling home page request",
		logger.String("path", c.Path()),
		logger.String("method", c.Request().Method),
	)
	return components.Home().Render(c.Request().Context(), c.Response().Writer)
}

func (h *MarketingHandler) HandleContact(c echo.Context) error {
	h.logger.Debug("Handling contact page request",
		logger.String("path", c.Path()),
		logger.String("method", c.Request().Method),
	)
	return components.Contact().Render(c.Request().Context(), c.Response().Writer)
}
