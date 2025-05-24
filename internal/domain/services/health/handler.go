package health

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Handler handles health check requests
type Handler struct {
	logger  logging.Logger
	service Service
}

// NewHandler creates a new health handler
func NewHandler(logger logging.Logger, service Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// Register registers the health check routes
func (h *Handler) Register(e *echo.Echo) {
	e.GET("/health", h.HandleHealthCheck)
	h.logger.Debug("registered health check endpoint",
		logging.StringField("path", "/health"),
		logging.StringField("method", "GET"),
	)
}

// HandleHealthCheck handles health check requests
func (h *Handler) HandleHealthCheck(c echo.Context) error {
	// Check system health
	status, err := h.service.CheckHealth(c.Request().Context())
	if err != nil {
		return response.ErrorResponse(c, http.StatusServiceUnavailable, "Service is not healthy: database connection failed")
	}

	// Return health status
	return response.Success(c, status)
}
