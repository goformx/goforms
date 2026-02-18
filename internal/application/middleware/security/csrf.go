// Package security provides security-related middleware configuration.
package security

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/goformx/goforms/internal/application/constants"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// SetupCSRF creates and configures CSRF middleware
func SetupCSRF(
	csrfConfig *appconfig.CSRFConfig,
	isDevelopment bool,
	logger logging.Logger,
) echo.MiddlewareFunc {
	sameSite := getSameSite(csrfConfig.CookieSameSite, isDevelopment)
	tokenLength := getTokenLength(csrfConfig.TokenLength)

	// Create component-scoped logger for CSRF
	csrfLogger := logger.WithComponent("csrf")

	// Log CSRF configuration at debug level
	if isDevelopment {
		csrfLogger.Debug("CSRF middleware configured",
			"context_key", csrfConfig.ContextKey,
			"token_lookup", csrfConfig.TokenLookup,
			"cookie_name", csrfConfig.CookieName,
			"cookie_secure", !isDevelopment,
			"cookie_same_site", sameSite,
		)
	}

	// Wrap Echo's CSRF middleware to add debug logging
	csrfMiddleware := echomw.CSRFWithConfig(echomw.CSRFConfig{
		TokenLength:    uint8(tokenLength), // #nosec G115
		TokenLookup:    csrfConfig.TokenLookup,
		ContextKey:     csrfConfig.ContextKey,
		CookieName:     csrfConfig.CookieName,
		CookiePath:     csrfConfig.CookiePath,
		CookieDomain:   csrfConfig.CookieDomain,
		CookieSecure:   !isDevelopment,
		CookieHTTPOnly: csrfConfig.CookieHTTPOnly,
		CookieSameSite: sameSite,
		CookieMaxAge:   csrfConfig.CookieMaxAge,
		Skipper:        CreateCSRFSkipper(isDevelopment, csrfLogger),
		ErrorHandler:   CreateCSRFErrorHandler(csrfConfig, isDevelopment, csrfLogger),
	})

	// Wrap middleware to add debug logging after it runs
	if isDevelopment {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			csrfHandler := csrfMiddleware(next)
			return func(c echo.Context) error {
				// Call the CSRF middleware
				err := csrfHandler(c)

				// After CSRF middleware runs, check if token exists
				if token, ok := c.Get(csrfConfig.ContextKey).(string); ok && token != "" {
					csrfLogger.Debug("CSRF token found in context",
						"token_length", len(token),
						"path", c.Request().URL.Path,
					)
				} else {
					csrfLogger.Debug("CSRF token not in context",
						"context_key", csrfConfig.ContextKey,
						"path", c.Request().URL.Path,
					)
				}

				if cookie, cookieErr := c.Cookie(csrfConfig.CookieName); cookieErr == nil && cookie != nil && cookie.Value != "" {
					csrfLogger.Debug("CSRF token found in cookie",
						"token_length", len(cookie.Value),
						"path", c.Request().URL.Path,
					)
				}

				return err
			}
		}
	}

	return csrfMiddleware
}

// getSameSite converts string SameSite to http.SameSite
func getSameSite(cookieSameSite string, isDevelopment bool) http.SameSite {
	switch cookieSameSite {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		if isDevelopment {
			return http.SameSiteLaxMode
		}
		return http.SameSiteStrictMode
	}
}

// getTokenLength ensures token length is within bounds for uint8
func getTokenLength(tokenLength int) int {
	if tokenLength <= 0 || tokenLength > 255 {
		return constants.DefaultTokenLength
	}
	return tokenLength
}

// CreateCSRFSkipper creates a CSRF skipper function
func CreateCSRFSkipper(isDevelopment bool, logger logging.Logger) func(c echo.Context) bool {
	return func(c echo.Context) bool {
		path := c.Request().URL.Path
		method := c.Request().Method

		if isDevelopment {
			logCSRFSkipperDebug(logger, path, method)
		}

		if IsSafeMethod(method) {
			skipResult := handleSafeMethodCSRF(logger, path, isDevelopment)
			if !skipResult {
				// handleSafeMethodCSRF explicitly said DON'T skip
				// This means it's a form/auth page that needs a token
				if isDevelopment {
					logger.Debug("CSRF skipper: form/auth page detected, forcing token generation",
						"path", path, "method", method)
				}
				return false // Don't skip, generate token
			}
			// If skipResult is true, continue to other checks
		}

		if shouldSkipCSRFForRoute(logger, path, isDevelopment) {
			return true
		}

		if isDevelopment {
			logger.Debug("CSRF not skipped - requires protection", "path", path, "method", method)
		}

		return false
	}
}

// logCSRFSkipperDebug logs debug information for CSRF skipper
func logCSRFSkipperDebug(logger logging.Logger, path, method string) {
	logger.Debug("CSRF skipper check",
		"path", path,
		"method", method,
		"is_safe_method", IsSafeMethod(method),
		"is_auth_page", IsAuthPage(path),
		"is_form_page", IsFormPage(path),
		"is_api_route", IsAPIRoute(path),
		"is_health_route", IsHealthRoute(path),
		"is_static_route", IsStaticRoute(path),
		"is_form_submission_route", IsFormSubmissionRoute(path),
		"is_auth_endpoint", IsAuthEndpoint(path))
}

// handleSafeMethodCSRF handles CSRF logic for safe HTTP methods
func handleSafeMethodCSRF(logger logging.Logger, path string, isDevelopment bool) bool {
	if IsAuthPage(path) || IsFormPage(path) {
		if isDevelopment {
			logger.Debug("CSRF not skipped - token generation needed", "path", path)
		}
		return false
	}

	return true
}

// shouldSkipCSRFForRoute checks if CSRF should be skipped for the given route
func shouldSkipCSRFForRoute(logger logging.Logger, path string, isDevelopment bool) bool {
	// Skip CSRF for public form submission endpoints first (before form page check)
	// These are API endpoints for embedded forms, not form builder pages
	if IsFormSubmissionRoute(path) {
		return true
	}

	// NEVER skip CSRF for form pages or auth pages - they ALWAYS need tokens
	// This acts as a safety guard even if other checks are misconfigured
	if IsFormPage(path) || IsAuthPage(path) {
		if isDevelopment {
			logger.Debug("CSRF skipper: never skipping form/auth pages", "path", path)
		}
		return false
	}

	if IsAuthEndpoint(path) {
		return true
	}

	if isDevelopment && IsAPIRoute(path) {
		return true
	}

	if IsHealthRoute(path) {
		return true
	}

	if IsStaticRoute(path) {
		return true
	}

	return false
}

// CreateCSRFErrorHandler creates the CSRF error handler function
func CreateCSRFErrorHandler(
	csrfConfig *appconfig.CSRFConfig,
	isDevelopment bool,
	logger logging.Logger,
) func(err error, c echo.Context) error {
	return func(err error, c echo.Context) error {
		if isDevelopment {
			csrfToken := c.Request().Header.Get("X-Csrf-Token")
			contextToken := ""
			if token, ok := c.Get(csrfConfig.ContextKey).(string); ok {
				contextToken = token
			}

			logger.Error("CSRF validation failed",
				"error", err.Error(),
				"path", c.Request().URL.Path,
				"method", c.Request().Method,
				"token_lookup", csrfConfig.TokenLookup,
				"origin", c.Request().Header.Get("Origin"),
				"csrf_token_present", csrfToken != "",
				"csrf_token_length", len(csrfToken),
				"context_token_present", contextToken != "",
				"context_token_length", len(contextToken),
				"content_type", c.Request().Header.Get("Content-Type"),
			)
		} else {
			// In production, log minimal info without sensitive details
			logger.Warn("CSRF validation failed",
				"path", c.Request().URL.Path,
				"method", c.Request().Method,
			)
		}

		return c.NoContent(http.StatusForbidden)
	}
}

// IsSafeMethod checks if the HTTP method is safe (doesn't modify state)
func IsSafeMethod(method string) bool {
	return method == "GET" || method == "HEAD" || method == "OPTIONS"
}

// IsAPIRoute checks if the path is an API route
func IsAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

// IsHealthRoute checks if the path is a health check route
func IsHealthRoute(path string) bool {
	return path == "/health" || path == "/health/" || path == "/healthz" || path == "/healthz/"
}

// IsStaticRoute checks if the path is a static asset route
func IsStaticRoute(path string) bool {
	return strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/public/") ||
		strings.HasPrefix(path, "/favicon.ico")
}

// IsFormSubmissionRoute checks if the path is a form submission endpoint
func IsFormSubmissionRoute(path string) bool {
	// Only skip CSRF for public form submission endpoints
	// (e.g., embedded form submissions from external sites)

	// Match /api/v1/forms/:id/submit
	if strings.HasPrefix(path, "/api/v1/forms/") && strings.HasSuffix(path, "/submit") {
		return true
	}

	// Match /forms/:id/submit (public embed routes)
	if strings.HasPrefix(path, "/forms/") && strings.HasSuffix(path, "/submit") {
		return true
	}

	// Check for direct submission endpoints
	if strings.HasPrefix(path, "/submit/") {
		return true
	}

	// Do NOT skip for:
	// - /forms/new (form creation page)
	// - /forms/:id/edit (form edit page)
	// These pages need CSRF tokens

	return false
}

// IsAuthPage checks if the path is an authentication page
func IsAuthPage(path string) bool {
	return path == "/login" || path == "/signup" ||
		path == "/forgot-password" || path == "/reset-password"
}

// IsAuthEndpoint checks if the path is an authentication endpoint
func IsAuthEndpoint(path string) bool {
	return path == "/login" || path == "/signup" || path == "/logout" ||
		path == "/forgot-password" || path == "/reset-password"
}

// IsFormPage checks if the path is a form page
func IsFormPage(path string) bool {
	return strings.Contains(path, "/forms/new") ||
		strings.Contains(path, "/forms/") || strings.Contains(path, "/submit") ||
		strings.Contains(path, "/dashboard")
}
