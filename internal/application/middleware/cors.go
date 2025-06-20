package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
)

// CORS creates and configures CORS middleware
func CORS(securityConfig *appconfig.SecurityConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method

			// Use general CORS settings for all endpoints
			// Form-specific CORS will be implemented when the dashboard is ready
			config := &corsConfig{
				allowedOrigins:   securityConfig.CorsAllowedOrigins,
				allowedMethods:   securityConfig.CorsAllowedMethods,
				allowedHeaders:   securityConfig.CorsAllowedHeaders,
				allowCredentials: securityConfig.CorsAllowCredentials,
				maxAge:           securityConfig.CorsMaxAge,
			}

			// Handle preflight requests
			if method == "OPTIONS" {
				return handlePreflightRequest(c, config)
			}

			// Handle actual requests
			return handleActualRequest(c, config, next)
		}
	}
}

// corsConfig holds CORS configuration
type corsConfig struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	allowCredentials bool
	maxAge           int
}

// handlePreflightRequest handles OPTIONS requests
func handlePreflightRequest(c echo.Context, config *corsConfig) error {
	origin := c.Request().Header.Get("Origin")

	if isOriginAllowed(origin, config.allowedOrigins) {
		c.Response().Header().Set("Access-Control-Allow-Origin", origin)
		c.Response().Header().Set("Access-Control-Allow-Methods", strings.Join(config.allowedMethods, ","))
		c.Response().Header().Set("Access-Control-Allow-Headers", strings.Join(config.allowedHeaders, ","))
		if config.allowCredentials {
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Response().Header().Set("Access-Control-Max-Age", strconv.Itoa(config.maxAge))
		return c.NoContent(http.StatusNoContent)
	}

	return c.NoContent(http.StatusNoContent)
}

// handleActualRequest handles actual requests (non-OPTIONS)
func handleActualRequest(c echo.Context, config *corsConfig, next echo.HandlerFunc) error {
	origin := c.Request().Header.Get("Origin")
	if origin != "" && isOriginAllowed(origin, config.allowedOrigins) {
		c.Response().Header().Set("Access-Control-Allow-Origin", origin)
		if config.allowCredentials {
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}

	return next(c)
}

// isOriginAllowed checks if the origin is in the allowed origins list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}
