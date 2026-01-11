package security

import (
	"github.com/labstack/echo/v4"

	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
)

// SetupSecurityHeaders creates middleware for additional security headers
func SetupSecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			securityConfig, ok := c.Get("security_config").(*appconfig.SecurityConfig)
			if !ok {
				// Fallback to default values if config not available
				c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
				c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			} else {
				c.Response().Header().Set("Referrer-Policy", securityConfig.SecurityHeaders.ReferrerPolicy)
				c.Response().Header().Set("Strict-Transport-Security", securityConfig.SecurityHeaders.StrictTransportSecurity)
				c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			}

			// Add exposed headers for CORS
			origin := c.Request().Header.Get("Origin")
			if origin != "" {
				c.Response().Header().Set("Access-Control-Expose-Headers", "X-Csrf-Token")
			}

			return next(c)
		}
	}
}

// IsNoisePath checks if the path should be suppressed from logging
func IsNoisePath(path string) bool {
	return path == "/.well-known" ||
		path == "/favicon.ico" ||
		path == "/robots.txt" ||
		containsDevTools(path)
}

func containsDevTools(path string) bool {
	return path == "com.chrome.devtools" ||
		path == "devtools" ||
		path == "chrome-devtools"
}
