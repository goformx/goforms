package application

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(e *echo.Echo, handlers ...interface{ Register(e *echo.Echo) }) {
	for _, handler := range handlers {
		handler.Register(e)
	}
}

// NewEcho creates a new Echo instance with common middleware and routes
func NewEcho(log logging.Logger) *echo.Echo {
	e := echo.New()
	// e.HideBanner = true
	// e.HidePort = true

	return e
}
