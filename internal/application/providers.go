package application

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(e *echo.Echo, handlers ...interface{ Register(e *echo.Echo) }) {
	// Register web routes (unprotected)
	e.Static("/static", "static")
	e.File("/favicon.ico", "static/favicon.ico")
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to GoForms API")
	})

	// Register API handlers
	for _, handler := range handlers {
		handler.Register(e)
	}
}

// NewEcho creates a new Echo instance with common middleware and routes
func NewEcho(log logging.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	return e
}
