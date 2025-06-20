package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	formdomain "github.com/goformx/goforms/internal/domain/form"
	formmodel "github.com/goformx/goforms/internal/domain/form/model"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// PerFormCORSConfig holds configuration for the per-form CORS middleware
type PerFormCORSConfig struct {
	FormService    formdomain.Service
	Logger         logging.Logger
	GlobalCORS     *appconfig.SecurityConfig
	FormRouteRegex *regexp.Regexp
}

// NewPerFormCORSConfig creates a new PerFormCORS configuration
func NewPerFormCORSConfig(
	formService formdomain.Service,
	logger logging.Logger,
	globalCORS *appconfig.SecurityConfig,
) *PerFormCORSConfig {
	// Regex to match form routes: /forms/:id or /api/v1/forms/:id
	formRouteRegex := regexp.MustCompile(`^/(?:forms|api/v1/forms)/([^/]+)(?:/.*)?$`)

	return &PerFormCORSConfig{
		FormService:    formService,
		Logger:         logger,
		GlobalCORS:     globalCORS,
		FormRouteRegex: formRouteRegex,
	}
}

// PerFormCORS creates middleware that applies form-specific CORS settings
func PerFormCORS(config *PerFormCORSConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if this is a form route
			formID := extractFormID(c.Request().URL.Path, config.FormRouteRegex)
			if formID == "" {
				// Not a form route, apply global CORS
				return applyGlobalCORS(c, config.GlobalCORS, next)
			}

			// Load form CORS configuration
			form, err := config.FormService.GetForm(c.Request().Context(), formID)
			if err != nil {
				config.Logger.Debug("failed to load form for CORS",
					"form_id", formID,
					"error", err,
					"falling_back_to_global_cors", true)
				// Fallback to global CORS
				return applyGlobalCORS(c, config.GlobalCORS, next)
			}

			if form == nil {
				config.Logger.Debug("form not found for CORS",
					"form_id", formID,
					"falling_back_to_global_cors", true)
				// Fallback to global CORS
				return applyGlobalCORS(c, config.GlobalCORS, next)
			}

			// Apply form-specific CORS
			return applyFormCORS(c, form, config.GlobalCORS, next)
		}
	}
}

// extractFormID extracts the form ID from the URL path
func extractFormID(path string, formRouteRegex *regexp.Regexp) string {
	matches := formRouteRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// applyFormCORS applies form-specific CORS headers
func applyFormCORS(c echo.Context, form *formmodel.Form, globalCORS *appconfig.SecurityConfig, next echo.HandlerFunc) error {
	// Get form CORS configuration
	origins, methods, headers := form.GetCorsConfig()

	// Use form CORS settings if available, otherwise fallback to global
	if len(origins) == 0 {
		origins = globalCORS.CorsAllowedOrigins
	}
	if len(methods) == 0 {
		methods = globalCORS.CorsAllowedMethods
	}
	if len(headers) == 0 {
		headers = globalCORS.CorsAllowedHeaders
	}

	// Handle preflight requests
	if c.Request().Method == http.MethodOptions {
		return handlePreflight(c, origins, methods, headers, globalCORS.CorsAllowCredentials, globalCORS.CorsMaxAge)
	}

	// Handle actual requests
	return handleActualRequest(c, origins, methods, headers, globalCORS.CorsAllowCredentials, next)
}

// applyGlobalCORS applies global CORS headers as fallback
func applyGlobalCORS(c echo.Context, globalCORS *appconfig.SecurityConfig, next echo.HandlerFunc) error {
	// Handle preflight requests
	if c.Request().Method == http.MethodOptions {
		return handlePreflight(
			c,
			globalCORS.CorsAllowedOrigins,
			globalCORS.CorsAllowedMethods,
			globalCORS.CorsAllowedHeaders,
			globalCORS.CorsAllowCredentials,
			globalCORS.CorsMaxAge,
		)
	}

	// Handle actual requests
	return handleActualRequest(
		c,
		globalCORS.CorsAllowedOrigins,
		globalCORS.CorsAllowedMethods,
		globalCORS.CorsAllowedHeaders,
		globalCORS.CorsAllowCredentials,
		next,
	)
}

// handlePreflight handles OPTIONS preflight requests
func handlePreflight(c echo.Context, origins, methods, headers []string, allowCredentials bool, maxAge int) error {
	origin := c.Request().Header.Get("Origin")

	// Check if origin is allowed
	if !isOriginAllowed(origin, origins) {
		return c.NoContent(http.StatusForbidden)
	}

	// Set CORS headers
	c.Response().Header().Set("Access-Control-Allow-Origin", origin)
	c.Response().Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	c.Response().Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	c.Response().Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", maxAge))

	if allowCredentials {
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
	}

	return c.NoContent(http.StatusNoContent)
}

// handleActualRequest handles actual CORS requests
func handleActualRequest(c echo.Context, origins, methods, headers []string, allowCredentials bool, next echo.HandlerFunc) error {
	origin := c.Request().Header.Get("Origin")

	// Check if origin is allowed
	if !isOriginAllowed(origin, origins) {
		return c.NoContent(http.StatusForbidden)
	}

	// Set CORS headers
	c.Response().Header().Set("Access-Control-Allow-Origin", origin)
	c.Response().Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	c.Response().Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))

	if allowCredentials {
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
	}

	return next(c)
}

// isOriginAllowed checks if the origin is allowed based on the CORS configuration
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return true // No origin header, allow
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true // Wildcard allows all origins
		}
		if allowed == origin {
			return true // Exact match
		}
		// TODO: Add support for pattern matching (e.g., *.example.com)
	}

	return false
}
