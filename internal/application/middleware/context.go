package middleware

import (
	"context"
	"time"

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
	// RequestIDKey is the context key for the request ID
	RequestIDKey contextKey = "request_id"
	// CorrelationIDKey is the context key for the correlation ID
	CorrelationIDKey contextKey = "correlation_id"
	// LoggerKey is the context key for the logger
	LoggerKey contextKey = "logger"
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UserRoleKey is the context key for user role
	UserRoleKey contextKey = "user_role"
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

// GetLogger retrieves the logger from context
func GetLogger(ctx context.Context) logging.Logger {
	if logger, ok := ctx.Value(LoggerKey).(logging.Logger); ok {
		return logger
	}
	return nil
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID retrieves the user ID from context
func GetUserID(c echo.Context) uint {
	userID, ok := c.Get(string(UserIDKey)).(uint)
	if !ok {
		return 0
	}

	return userID
}

// GetUserRole retrieves the user role from context
func GetUserRole(c echo.Context) string {
	role, ok := c.Get(string(UserRoleKey)).(string)
	if !ok {
		return ""
	}

	return role
}
