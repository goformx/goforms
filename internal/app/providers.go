package app

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jonesrussell/goforms/internal/logger"
)

// NewEcho creates a new Echo instance with common middleware
func NewEcho(log logger.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Configure static file serving with proper caching and security
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "static",
		Browse: false,
		HTML5:  true,
		Index:  "",
		// Add security headers
		Skipper: func(c echo.Context) bool {
			return !strings.HasPrefix(c.Path(), "/static")
		},
	}))

	// Add cache control headers for static files
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Path(), "/static") {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000")
			}
			return next(c)
		}
	})

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	return e
}
