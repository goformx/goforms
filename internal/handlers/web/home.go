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
		BaseHandler: handlers.NewBaseHandler(nil, logger),
	}
}

// Register registers the home routes
func (h *HomeHandler) Register(e *echo.Echo) {
	e.GET("/", h.Home)
}

// Home handles the home page request
func (h *HomeHandler) Home(c echo.Context) error {
	// Check if user is authenticated
	if user := c.Get("user"); user != nil {
		// User is authenticated, redirect to dashboard
		return c.Redirect(http.StatusFound, "/dashboard")
	}

	// User is not authenticated, redirect to login
	return c.Redirect(http.StatusFound, "/login")
}
