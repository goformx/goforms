package middleware

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/domain/common/ctxutil"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// contextKey is a custom type for context keys
type contextKey string

const (
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

// WithContext adds context handling to the request
func (m *ContextMiddleware) WithContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
			defer cancel()

			// Add request ID to context
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = c.Response().Header().Get("X-Request-ID")
			}
			ctx = ctxutil.WithRequestID(ctx, requestID)

			// Add trace ID to context
			traceID := c.Request().Header.Get("X-Trace-ID")
			if traceID != "" {
				ctx = ctxutil.WithTraceID(ctx, traceID)
			}

			// Add correlation ID to context
			correlationID := c.Request().Header.Get("X-Correlation-ID")
			if correlationID != "" {
				ctx = ctxutil.WithCorrelationID(ctx, correlationID)
			}

			// Add user ID to context if available
			if userID := c.Get("user_id"); userID != nil {
				if id, ok := userID.(uint); ok {
					ctx = ctxutil.WithUserID(ctx, id)
				}
			}

			// Create a logger with context
			logger := m.logger.With(
				logging.String("request_id", requestID),
				logging.String("method", c.Request().Method),
				logging.String("path", c.Request().URL.Path),
				logging.String("remote_addr", c.Request().RemoteAddr),
				logging.String("user_agent", c.Request().UserAgent()),
			)

			// Add trace and correlation IDs if available
			if traceID != "" {
				logger = logger.With(logging.String("trace_id", traceID))
			}
			if correlationID != "" {
				logger = logger.With(logging.String("correlation_id", correlationID))
			}

			// Add logger to context
			ctx = context.WithValue(ctx, LoggerKey, logger)

			// Update request context
			c.SetRequest(c.Request().WithContext(ctx))

			// Log request start with timing
			startTime := time.Now()
			logger.Info("Request started")

			// Handle request
			err := next(c)

			// Calculate request duration
			duration := time.Since(startTime)

			// Log request end with timing and status
			if err != nil {
				logger.Error("Request failed",
					logging.Error(err),
					logging.Duration("duration", duration),
					logging.Int("status", c.Response().Status),
				)
			} else {
				logger.Info("Request completed",
					logging.Duration("duration", duration),
					logging.Int("status", c.Response().Status),
				)
			}

			return err
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
