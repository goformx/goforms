// Package health provides HTTP handlers for health checks and application status endpoints.
package health

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/health"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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
		h.logger.Error("health check failed", "error", err)
		if jsonErr := c.JSON(http.StatusServiceUnavailable, status); jsonErr != nil {
			return fmt.Errorf("return health check error response: %w", jsonErr)
		}
		return nil
	}

	if jsonErr := c.JSON(http.StatusOK, status); jsonErr != nil {
		return fmt.Errorf("return health check success response: %w", jsonErr)
	}
	return nil
}
