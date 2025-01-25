package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CSRFMiddleware creates a new CSRF middleware
func CSRFMiddleware() echo.MiddlewareFunc {
	config := middleware.CSRFConfig{
		TokenLookup: "header:X-CSRF-Token",
	}
	return middleware.CSRFWithConfig(config)
}
