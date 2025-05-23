package services

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// PingContexter is an interface for database health checks
type PingContexter interface {
	PingContext(ctx echo.Context) error
}

// HealthHandler handles health check requests
type HealthHandler struct {
	logger logging.Logger
	db     PingContexter
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(log logging.Logger, db PingContexter) *HealthHandler {
	return &HealthHandler{
		logger: log,
		db:     db,
	}
}

// Register registers the health check routes
func (h *HealthHandler) Register(e *echo.Echo) {
	e.GET("/health", h.HandleHealthCheck)
}

// HandleHealthCheck handles health check requests
func (h *HealthHandler) HandleHealthCheck(c echo.Context) error {
	if err := h.db.PingContext(c); err != nil {
		h.logger.Error("health check failed", logging.ErrorField("error", err))
		return response.InternalError(c, "Service is not healthy")
	}

	if err := response.Success(c, map[string]any{
		"status": "healthy",
	}); err != nil {
		h.logger.Error("failed to send health check response", logging.ErrorField("error", err))
		return fmt.Errorf("failed to send health check response: %w", err)
	}

	return nil
}
