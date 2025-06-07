package middleware

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/domain/common/ctxutil"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	// RequestIDHeader is the canonical header name for request ID
	RequestIDHeader = "X-Trace-Id"
	// RequestTimeout is the default timeout for requests
	RequestTimeout = 30 * time.Second
)

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
	// LoggerKey is the context key for the logger
	LoggerKey contextKey = "logger"
)

// ContextMiddleware provides context handling for HTTP requests
type ContextMiddleware struct {
	logger logging.Logger
}

// NewContextMiddleware creates a new context middleware
func NewContextMiddleware(logger logging.Logger) *ContextMiddleware {
	return &ContextMiddleware{
		logger: logger,
	}
}

// WithContext adds context to the request
func (m *ContextMiddleware) WithContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get or generate request ID
			requestID := c.Request().Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
				c.Request().Header.Set(RequestIDHeader, requestID)
			}

			// Create request context with timeout
			ctx, cancel := context.WithTimeout(c.Request().Context(), RequestTimeout)
			defer cancel()

			// Add request ID and logger to context
			ctx = context.WithValue(ctx, RequestIDKey, requestID)
			ctx = context.WithValue(ctx, LoggerKey, m.logger)

			// Update request context
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// GetLoggerFromContext retrieves the logger from context
func GetLoggerFromContext(ctx context.Context) logging.Logger {
	if logger, ok := ctx.Value(LoggerKey).(logging.Logger); ok {
		return logger
	}
	return nil
}

// GetRequestIDFromContext retrieves the request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctxutil.GetRequestID(ctx); ok {
		return requestID
	}
	return ""
}

// GetUserIDFromContext retrieves the user ID from context
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	return ctxutil.GetUserID(ctx)
}

// GetTraceIDFromContext retrieves the trace ID from context
func GetTraceIDFromContext(ctx context.Context) (string, bool) {
	return ctxutil.GetTraceID(ctx)
}

// GetCorrelationIDFromContext retrieves the correlation ID from context
func GetCorrelationIDFromContext(ctx context.Context) (string, bool) {
	return ctxutil.GetCorrelationID(ctx)
}
