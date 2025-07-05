package health

import (
	"net/http"

	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
)

// HealthHandler handles the /health endpoint
// Implements httpiface.Handler
type HealthHandler struct {
	handlers.BaseHandler
}

// NewHealthHandler creates a new HealthHandler and registers the /health route
func NewHealthHandler() *HealthHandler {
	h := &HealthHandler{
		BaseHandler: *handlers.NewBaseHandler("health"),
	}

	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/health",
		Handler: h.handleHealth,
	})

	return h
}

// handleHealth serves the /health endpoint
func (h *HealthHandler) handleHealth(ctx httpiface.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   "now", // You can add a timestamp if needed
	})
}
