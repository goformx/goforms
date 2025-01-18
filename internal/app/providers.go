package app

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// NewEcho creates a new Echo instance with common middleware
func NewEcho(logger *zap.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static files
	e.Static("/static", "static")

	return e
}
