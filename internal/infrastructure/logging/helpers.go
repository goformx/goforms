package logging

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// ContextKey represents a key in the context for logging purposes.
type ContextKey string

const (
	// LoggerContextKey is the context key for logger.
	LoggerContextKey ContextKey = "logger"
	// RequestIDContextKey is the context key for request ID.
	RequestIDContextKey ContextKey = "request_id"
	// FormIDContextKey is the context key for form ID.
	FormIDContextKey ContextKey = "form_id"
	// UserIDContextKey is the context key for user ID.
	UserIDContextKey ContextKey = "user_id"
)

// defaultLogger is the fallback logger when context doesn't have one.
var defaultLogger Logger

// SetDefaultLogger sets the default fallback logger.
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// LoggerFromContext extracts the logger from context and enriches it with
// available request context fields (request_id, form_id, user_id).
// If no logger is found in context, returns the default logger.
func LoggerFromContext(ctx context.Context) Logger {
	logger := getLoggerFromContext(ctx)
	if logger == nil {
		if defaultLogger != nil {
			logger = defaultLogger
		} else {
			// Return a no-op logger if nothing is available
			return nil
		}
	}

	// Auto-enrich with available context values
	return enrichLoggerFromContext(logger, ctx)
}

// getLoggerFromContext retrieves the raw logger from context.
func getLoggerFromContext(ctx context.Context) Logger {
	// Try our logging context key first
	if logger, ok := ctx.Value(LoggerContextKey).(Logger); ok {
		return logger
	}

	// Try the middleware context key (using string comparison for loose coupling)
	if logger, ok := ctx.Value("logger").(Logger); ok {
		return logger
	}

	return nil
}

// enrichLoggerFromContext adds context fields to the logger.
func enrichLoggerFromContext(logger Logger, ctx context.Context) Logger {
	// Add request ID if present
	if reqID := getStringFromContext(ctx, RequestIDContextKey, "request_id"); reqID != "" {
		logger = logger.WithRequestID(reqID)
	}

	// Add form ID if present
	if formID := getStringFromContext(ctx, FormIDContextKey, "form_id"); formID != "" {
		logger = logger.With("form_id", formID)
	}

	// Add user ID if present
	if userID := getStringFromContext(ctx, UserIDContextKey, "user_id"); userID != "" {
		logger = logger.WithUserID(userID)
	}

	return logger
}

// getStringFromContext tries multiple keys to get a string value from context.
func getStringFromContext(ctx context.Context, keys ...any) string {
	for _, key := range keys {
		if value, ok := ctx.Value(key).(string); ok && value != "" {
			return value
		}
	}

	return ""
}

// WithContext returns a new context with the logger attached.
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}

// convertToZapFields converts a slice of fields to zap fields with performance optimization
func convertToZapFields(fields []any, fieldSanitizer *Sanitizer) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)/FieldPairSize)

	for i := 0; i < len(fields); i += FieldPairSize {
		if i+1 >= len(fields) {
			break
		}

		key, ok := fields[i].(string)
		if !ok {
			continue
		}

		value := fields[i+1]

		// Use optimized field creation that preserves native types
		zapFields = append(zapFields, createOptimizedField(key, value, fieldSanitizer))
	}

	return zapFields
}

// createOptimizedField creates a zap field with type preservation and selective sanitization
func createOptimizedField(key string, value any, fieldSanitizer *Sanitizer) zap.Field {
	// Check if this is a sensitive field first
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	// Preserve native types when possible
	switch v := value.(type) {
	case string:
		// Only sanitize strings that need it
		if needsStringSanitization(key, v) {
			return zap.String(key, fieldSanitizer.SanitizeField(key, v))
		}

		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case int64:
		return zap.Int64(key, v)
	case float64:
		return zap.Float64(key, v)
	case bool:
		return zap.Bool(key, v)
	case error:
		return zap.Error(v)
	default:
		// For complex types, use sanitization
		return zap.String(key, fieldSanitizer.SanitizeField(key, value))
	}
}

// needsStringSanitization determines if a string field needs sanitization
func needsStringSanitization(key, value string) bool {
	// Only sanitize strings that might contain dangerous content
	switch key {
	case "path", "file_path", "url", "user_agent", "referer", "origin":
		return true
	case "error", "err", "message", "msg":
		return true
	default:
		// Check for dangerous characters in any string
		return strings.ContainsAny(value, "\n\r\x00<>\"'\\")
	}
}

// sanitizeError sanitizes an error for safe logging
func sanitizeError(err error, sanitizer sanitization.ServiceInterface) string {
	if err == nil {
		return ""
	}

	// Get the error message and sanitize it
	errMsg := err.Error()

	// Apply the same sanitization as regular messages
	if sanitizer != nil {
		return sanitizer.SanitizeForLogging(errMsg)
	}

	return errMsg
}
