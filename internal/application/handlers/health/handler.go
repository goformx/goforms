package health

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/services/health"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// Handler handles health check requests
type Handler struct {
	service health.Service
	logger  logging.Logger
}

// NewHandler creates a new health check handler
func NewHandler(service health.Service, logger logging.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register registers the health check routes
func (h *Handler) Register(e *echo.Echo) {
	e.GET("/health", h.handleHealthCheck)
}

// handleHealthCheck handles the health check request
func (h *Handler) handleHealthCheck(c echo.Context) error {
	status, err := h.service.CheckHealth(c.Request().Context())
	if err != nil {
		h.logger.Error("health check failed", logging.ErrorField("error", err))
		return c.JSON(http.StatusServiceUnavailable, status)
	}

	return c.JSON(http.StatusOK, status)
}
