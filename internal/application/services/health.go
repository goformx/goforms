package services

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/application/response"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
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
		h.logger.Error("health check failed", logging.Error(err))
		return response.InternalError(c, "Service is not healthy")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
