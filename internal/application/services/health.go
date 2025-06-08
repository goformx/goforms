package services

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// PingContexter is an interface for database health checks
type PingContexter interface {
	PingContext(ctx context.Context) error
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
	h.logger.Debug("registered health check endpoint",
		logging.String("path", "/health"),
		logging.String("method", "GET"),
	)
}

// HandleHealthCheck handles health check requests
func (h *HealthHandler) HandleHealthCheck(c echo.Context) error {
	// Check database connectivity
	if err := h.db.PingContext(c.Request().Context()); err != nil {
		h.logger.Error("health check failed",
			logging.Error(err),
			logging.String("component", "database"),
		)
		return response.ErrorResponse(c, http.StatusServiceUnavailable, "Service is not healthy: database connection failed")
	}

	// Return health status
	return response.Success(c, map[string]any{
		"status": "healthy",
		"components": map[string]string{
			"database": "up",
		},
	})
}
