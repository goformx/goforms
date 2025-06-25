// Package health provides HTTP handlers for health checks and application status endpoints.
package health

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
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
		return response.ErrorResponse(c, http.StatusServiceUnavailable, "Health check failed")
	}

	return response.Success(c, status)
}
