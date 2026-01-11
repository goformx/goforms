// Package middleware provides HTTP middleware components for the application.
package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// DefaultSlowRequestThreshold is the default threshold for slow request detection.
	DefaultSlowRequestThreshold = 500 * time.Millisecond
	// VerySlowRequestThreshold is the threshold for very slow requests (logged as error).
	VerySlowRequestThreshold = 2 * time.Second
)

// SlowRequestConfig holds configuration for slow request detection middleware.
type SlowRequestConfig struct {
	// Threshold is the duration after which a request is considered slow.
	Threshold time.Duration
	// VerySlowThreshold is the duration after which a request is considered very slow.
	VerySlowThreshold time.Duration
	// Skipper defines a function to skip middleware for certain requests.
	Skipper func(c echo.Context) bool
}

// DefaultSlowRequestConfig returns default configuration for slow request detection.
func DefaultSlowRequestConfig() SlowRequestConfig {
	return SlowRequestConfig{
		Threshold:         DefaultSlowRequestThreshold,
		VerySlowThreshold: VerySlowRequestThreshold,
		Skipper:           nil,
	}
}

// SlowRequestDetector creates middleware that detects and logs slow requests.
func SlowRequestDetector(logger logging.Logger, threshold time.Duration) echo.MiddlewareFunc {
	config := DefaultSlowRequestConfig()
	config.Threshold = threshold

	return SlowRequestDetectorWithConfig(logger, config)
}

// SlowRequestDetectorWithConfig creates middleware with custom configuration.
func SlowRequestDetectorWithConfig(logger logging.Logger, config SlowRequestConfig) echo.MiddlewareFunc {
	slowLogger := logger.WithComponent("slow_request")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check skipper
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			// Only log if request exceeds threshold
			if duration > config.Threshold {
				logSlowRequest(slowLogger, c, duration, config)
			}

			return err
		}
	}
}

// logSlowRequest logs a slow request with appropriate level based on duration.
func logSlowRequest(logger logging.Logger, c echo.Context, duration time.Duration, config SlowRequestConfig) {
	fields := []any{
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"duration_ms", duration.Milliseconds(),
		"threshold_ms", config.Threshold.Milliseconds(),
		"status", c.Response().Status,
		"remote_ip", c.RealIP(),
	}

	// Add request ID if present
	if reqID := c.Request().Header.Get("X-Request-ID"); reqID != "" {
		fields = append(fields, "request_id", reqID)
	}

	// Add query parameters if present (but not values for security)
	if query := c.Request().URL.RawQuery; query != "" {
		fields = append(fields, "has_query", true)
	}

	// Very slow requests are logged as errors
	if duration > config.VerySlowThreshold {
		fields = append(fields, "severity", "very_slow")
		logger.Error("very slow request detected", fields...)

		return
	}

	// Slow requests are logged as warnings
	fields = append(fields, "severity", "slow")
	logger.Warn("slow request detected", fields...)
}

// NewSlowRequestSkipper creates a skipper that skips static assets and health checks.
func NewSlowRequestSkipper() func(c echo.Context) bool {
	return func(c echo.Context) bool {
		path := c.Request().URL.Path

		// Skip static assets
		if len(path) > 8 && path[:8] == "/assets/" {
			return true
		}
		if len(path) > 8 && path[:8] == "/static/" {
			return true
		}

		// Skip health checks
		if path == "/health" || path == "/healthz" {
			return true
		}

		return false
	}
}
