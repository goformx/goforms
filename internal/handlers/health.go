package handlers

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type HealthHandler struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sqlx.DB, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

func (h *HealthHandler) Register(e *echo.Echo) {
	e.GET("/health", h.Check)
}

func (h *HealthHandler) Check(c echo.Context) error {
	health := struct {
		Status    string `json:"status"`
		DBStatus  string `json:"db_status"`
		Timestamp string `json:"timestamp"`
	}{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := h.db.PingContext(c.Request().Context()); err != nil {
		h.logger.Error("database health check failed",
			zap.Error(err),
			zap.String("status", "degraded"),
		)
		health.DBStatus = "error"
		health.Status = "degraded"
		return c.JSON(http.StatusServiceUnavailable, health)
	}

	h.logger.Info("health check successful",
		zap.String("status", "ok"),
		zap.String("db_status", "ok"),
	)
	health.DBStatus = "ok"
	return c.JSON(http.StatusOK, health)
}
