package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/labstack/echo/v4"
)

// HomeHandler handles home page requests
type HomeHandler struct {
	*handlers.BaseHandler
}

// NewHomeHandler creates a new home handler
func NewHomeHandler(logger logging.Logger) *HomeHandler {
	return &HomeHandler{
		BaseHandler: handlers.NewBaseHandler(nil, nil, logger),
	}
}

// Register registers the home routes
func (h *HomeHandler) Register(e *echo.Echo) {
	e.GET("/", h.Home)
}

// Home handles the home page request
func (h *HomeHandler) Home(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/dashboard")
}
