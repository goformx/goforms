package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	formdomain "github.com/goformx/goforms/internal/domain/form"
	formmodel "github.com/goformx/goforms/internal/domain/form/model"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

const (
	formIDMatchIndex = 2
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
			formID := ExtractFormID(c.Request().URL.Path, config.FormRouteRegex)
			if formID == "" {
				// Not a form route, apply global CORS
				return applyGlobalCORS(c, config.GlobalCORS, next)
			}

			// Load form CORS configuration
			form, err := config.FormService.GetForm(c.Request().Context(), formID)
			if err != nil {
				config.Logger.Debug(
					"failed to load form for CORS",
					"form_id", config.Logger.SanitizeField("form_id", formID),
					"error", err,
					"falling_back_to_global_cors", true,
				)
				// Fallback to global CORS
				return applyGlobalCORS(c, config.GlobalCORS, next)
			}

			if form == nil {
				config.Logger.Debug(
					"form not found for CORS",
					"form_id", config.Logger.SanitizeField("form_id", formID),
					"falling_back_to_global_cors", true,
				)
				// Fallback to global CORS
				return applyGlobalCORS(c, config.GlobalCORS, next)
			}

			// Apply form-specific CORS
			return applyFormCORS(c, form, config.GlobalCORS, next)
		}
	}
}

// ExtractFormID extracts the form ID from the URL path
func ExtractFormID(path string, formRouteRegex *regexp.Regexp) string {
	matches := formRouteRegex.FindStringSubmatch(path)
	if len(matches) < formIDMatchIndex {
		return ""
	}
	return matches[1]
}

// applyFormCORS applies form-specific CORS headers
func applyFormCORS(
	c echo.Context,
	form *formmodel.Form,
	globalCORS *appconfig.SecurityConfig,
	next echo.HandlerFunc,
) error {
	// Get form CORS configuration
	origins, methods, headers := form.GetCorsConfig()

	// Use form CORS settings if available, otherwise fallback to global
	if len(origins) == 0 {
		origins = globalCORS.CORS.AllowedOrigins
	}
	if len(methods) == 0 {
		methods = globalCORS.CORS.AllowedMethods
	}
	if len(headers) == 0 {
		headers = globalCORS.CORS.AllowedHeaders
	}

	// Handle preflight requests
	if c.Request().Method == http.MethodOptions {
		return handlePreflight(
			c, origins, methods, headers,
			globalCORS.CORS.AllowCredentials, globalCORS.CORS.MaxAge,
		)
	}

	// Handle actual requests
	return handleActualRequest(
		c, origins, methods, headers,
		globalCORS.CORS.AllowCredentials, next,
	)
}

// applyGlobalCORS applies global CORS headers as fallback
func applyGlobalCORS(
	c echo.Context,
	globalCORS *appconfig.SecurityConfig,
	next echo.HandlerFunc,
) error {
	// Add debug logging
	if c.Logger() != nil {
		c.Logger().Debug("PerFormCORS: applying global CORS",
			"path", c.Request().URL.Path,
			"method", c.Request().Method,
			"origin", c.Request().Header.Get("Origin"),
			"allowed_origins", globalCORS.CORS.AllowedOrigins)
	}

	// Handle preflight requests
	if c.Request().Method == http.MethodOptions {
		return handlePreflight(
			c,
			globalCORS.CORS.AllowedOrigins,
			globalCORS.CORS.AllowedMethods,
			globalCORS.CORS.AllowedHeaders,
			globalCORS.CORS.AllowCredentials,
			globalCORS.CORS.MaxAge,
		)
	}

	// Handle actual requests
	return handleActualRequest(
		c,
		globalCORS.CORS.AllowedOrigins,
		globalCORS.CORS.AllowedMethods,
		globalCORS.CORS.AllowedHeaders,
		globalCORS.CORS.AllowCredentials,
		next,
	)
}

// handlePreflight handles OPTIONS preflight requests
func handlePreflight(
	c echo.Context,
	origins, methods, headers []string,
	allowCredentials bool,
	maxAge int,
) error {
	origin := c.Request().Header.Get("Origin")

	// Check if origin is allowed
	if !IsOriginAllowed(origin, origins) {
		return c.NoContent(http.StatusForbidden)
	}

	// Set CORS headers
	c.Response().Header().Set("Access-Control-Allow-Origin", origin)
	c.Response().Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	c.Response().Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	c.Response().Header().Set("Access-Control-Max-Age", strconv.Itoa(maxAge))

	if allowCredentials {
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
	}

	return c.NoContent(http.StatusNoContent)
}

// handleActualRequest handles actual CORS requests
func handleActualRequest(
	c echo.Context,
	origins, methods, headers []string,
	allowCredentials bool,
	next echo.HandlerFunc,
) error {
	origin := c.Request().Header.Get("Origin")

	// Check if origin is allowed
	if !IsOriginAllowed(origin, origins) {
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

// IsOriginAllowed checks if the origin is allowed based on the CORS configuration
func IsOriginAllowed(origin string, allowedOrigins []string) bool {
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
