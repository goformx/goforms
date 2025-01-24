package application

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(e *echo.Echo, handlers ...interface{ Register(e *echo.Echo) }) {
	// Register web routes (unprotected)
	e.Static("/static", "static")
	e.File("/favicon.ico", "static/favicon.ico")
	e.GET("/", func(c echo.Context) error {
		return pages.Home().Render(c.Request().Context(), c.Response().Writer)
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

	// Setup middleware
	mw := middleware.New(log)
	mw.Setup(e)

	return e
}
