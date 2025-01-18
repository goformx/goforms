package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/response"
	"github.com/labstack/echo/v4"
)

// PingContexter is an interface for database health checks
type PingContexter interface {
	PingContext(ctx echo.Context) error
}

// HealthHandler handles health check requests
type HealthHandler struct {
	logger logger.Logger
	db     PingContexter
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(log logger.Logger, db PingContexter) *HealthHandler {
	return &HealthHandler{
		logger: log,
		db:     db,
	}
}

// Register registers the health check routes
func (h *HealthHandler) Register(e *echo.Echo) {
	e.GET("/health", h.HandleHealth)
}

// HandleHealth handles health check requests
func (h *HealthHandler) HandleHealth(c echo.Context) error {
	h.logger.Debug("handling health check request")

	if err := h.db.PingContext(c); err != nil {
		h.logger.Error("database health check failed",
			logger.Error(err),
		)
		return response.Error(c, http.StatusServiceUnavailable, "database connection failed")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
