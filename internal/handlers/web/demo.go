package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/labstack/echo/v4"
)

// DemoHandler handles demo requests
type DemoHandler struct {
	*handlers.BaseHandler
}

// NewDemoHandler creates a new demo handler
func NewDemoHandler(logger logging.Logger) *DemoHandler {
	return &DemoHandler{
		BaseHandler: handlers.NewBaseHandler(nil, logger),
	}
}

// Register registers the demo routes
func (h *DemoHandler) Register(e *echo.Echo) {
	e.GET("/demo", h.Demo)
}

// Demo handles the demo page request
func (h *DemoHandler) Demo(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Demo"})
}
