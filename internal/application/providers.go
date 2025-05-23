package application

import (
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(e *echo.Echo, handlers ...interface{ Register(e *echo.Echo) }) {
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
	mw := middleware.New(&middleware.ManagerConfig{
		Logger: log,
	})
	mw.Setup(e)

	return e
}
