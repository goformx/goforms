package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure/config"
)

var defaultFormCORSMethods = []string{http.MethodGet, http.MethodPost, http.MethodOptions}
var defaultFormCORSHeaders = []string{"Content-Type", "Accept", "Origin"}

// NewFormCORSMiddleware enforces per-form CORS rules for public endpoints.
func NewFormCORSMiddleware(formService formdomain.Service, corsConfig config.CORSConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isPublicFormCORSRequest(c.Request().Method, c.Request().URL.Path) {
				return next(c)
			}

			origin := c.Request().Header.Get("Origin")
			if origin == "" {
				return next(c)
			}

			formID := extractFormIDFromPath(c.Request().URL.Path)
			if formID == "" {
				return next(c)
			}

			form, err := formService.GetForm(c.Request().Context(), formID)
			if err != nil || form == nil {
				return next(c)
			}

			allowedOrigins, allowedMethods, allowedHeaders := form.GetCorsConfig()
			if !isOriginAllowed(origin, allowedOrigins) {
				return c.NoContent(constants.StatusForbidden)
			}

			resolvedMethods := resolveCORSList(allowedMethods, corsConfig.AllowedMethods, defaultFormCORSMethods)
			resolvedHeaders := resolveCORSList(allowedHeaders, corsConfig.AllowedHeaders, defaultFormCORSHeaders)

			applyFormCORSHeaders(c, origin, resolvedMethods, resolvedHeaders, corsConfig)

			if c.Request().Method == http.MethodOptions {
				return c.NoContent(constants.StatusNoContent)
			}

			return next(c)
		}
	}
}

func resolveCORSList(list, fallback, defaultValues []string) []string {
	if len(list) > 0 {
		return list
	}
	if len(fallback) > 0 {
		return fallback
	}
	return defaultValues
}

func applyFormCORSHeaders(
	c echo.Context,
	origin string,
	allowedMethods []string,
	allowedHeaders []string,
	corsConfig config.CORSConfig,
) {
	headers := c.Response().Header()
	headers.Set("Access-Control-Allow-Origin", origin)
	headers.Set("Vary", "Origin")
	headers.Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
	headers.Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

	if corsConfig.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}

	if len(corsConfig.ExposedHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(corsConfig.ExposedHeaders, ", "))
	}

	if corsConfig.MaxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}

func isPublicFormCORSRequest(method, requestPath string) bool {
	prefix := constants.PathAPIForms + "/"
	publicPrefix := constants.PathFormsPublic + "/"
	if !strings.HasPrefix(requestPath, prefix) && !strings.HasPrefix(requestPath, publicPrefix) {
		return false
	}

	switch {
	case strings.HasSuffix(requestPath, "/schema"):
		return method == http.MethodGet || method == http.MethodOptions
	case strings.HasSuffix(requestPath, "/validation"):
		return method == http.MethodGet || method == http.MethodOptions
	case strings.HasSuffix(requestPath, "/submit"):
		return method == http.MethodPost || method == http.MethodOptions
	case strings.HasSuffix(requestPath, "/embed"):
		return method == http.MethodGet || method == http.MethodOptions
	default:
		return false
	}
}

func extractFormIDFromPath(requestPath string) string {
	prefix := constants.PathAPIForms + "/"
	publicPrefix := constants.PathFormsPublic + "/"
	var trimmed string
	if strings.HasPrefix(requestPath, prefix) {
		trimmed = strings.TrimPrefix(requestPath, prefix)
	} else if strings.HasPrefix(requestPath, publicPrefix) {
		trimmed = strings.TrimPrefix(requestPath, publicPrefix)
	} else {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}
