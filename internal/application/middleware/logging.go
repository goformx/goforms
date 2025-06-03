package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Logging is middleware that logs requests
type Logging struct {
	logger logging.Logger
}

// NewLogging creates a new logging middleware
func NewLogging(logger logging.Logger) *Logging {
	return &Logging{
		logger: logger,
	}
}

// Handle logs requests and responses using Echo's RequestLogger
func (m *Logging) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	config := middleware.RequestLoggerConfig{
		// Skip logging for static assets
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			return strings.HasPrefix(path, "/node_modules/") ||
				strings.HasPrefix(path, "/dist/") ||
				strings.HasPrefix(path, "/public/")
		},
		// Log all relevant request/response information
		LogLatency:       true,
		LogProtocol:      true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogURIPath:       true,
		LogRoutePath:     true,
		LogRequestID:     true,
		LogReferer:       true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogError:         true,
		LogContentLength: true,
		LogResponseSize:  true,
		// Log common headers
		LogHeaders: []string{
			"Content-Type",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Cache-Control",
			"Connection",
			"Cookie",
			"Origin",
			"Sec-Fetch-Dest",
			"Sec-Fetch-Mode",
			"Sec-Fetch-Site",
			"Upgrade-Insecure-Requests",
		},
		// Log common query parameters
		LogQueryParams: []string{
			"page",
			"limit",
			"sort",
			"filter",
			"search",
		},
		// Log common form values
		LogFormValues: []string{
			"email",
			"username",
			"password",
			"action",
		},
		// Custom logging function using our logger
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// Build log fields
			fields := []any{
				logging.StringField("method", v.Method),
				logging.StringField("uri", v.URI),
				logging.StringField("path", v.URIPath),
				logging.StringField("route", v.RoutePath),
				logging.StringField("remote_ip", v.RemoteIP),
				logging.StringField("host", v.Host),
				logging.StringField("protocol", v.Protocol),
				logging.StringField("request_id", v.RequestID),
				logging.StringField("referer", v.Referer),
				logging.StringField("user_agent", v.UserAgent),
				logging.IntField("status", v.Status),
				logging.AnyField("content_length", v.ContentLength),
				logging.AnyField("response_size", v.ResponseSize),
				logging.DurationField("latency", v.Latency),
			}

			// Add headers if present
			for k, v := range v.Headers {
				fields = append(fields, logging.StringField("header_"+strings.ToLower(k), strings.Join(v, ",")))
			}

			// Add query parameters if present
			for k, v := range v.QueryParams {
				fields = append(fields, logging.StringField("query_"+k, strings.Join(v, ",")))
			}

			// Add form values if present
			for k, v := range v.FormValues {
				fields = append(fields, logging.StringField("form_"+k, strings.Join(v, ",")))
			}

			// Log based on status code
			if v.Status >= 500 {
				m.logger.Error("request failed", fields...)
			} else if v.Status >= 400 {
				m.logger.Warn("request failed", fields...)
			} else {
				m.logger.Info("request completed", fields...)
			}

			return nil
		},
	}

	return middleware.RequestLoggerWithConfig(config)(next)
}
