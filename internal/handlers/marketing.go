package handlers

import (
	"github.com/jonesrussell/goforms/internal/components"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type MarketingHandler struct {
	logger *zap.Logger
}

func NewMarketingHandler(logger *zap.Logger) *MarketingHandler {
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
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method),
	)
	return components.Home().Render(c.Request().Context(), c.Response().Writer)
}

func (h *MarketingHandler) HandleContact(c echo.Context) error {
	h.logger.Debug("Handling contact page request",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method),
	)
	return components.Contact().Render(c.Request().Context(), c.Response().Writer)
}
