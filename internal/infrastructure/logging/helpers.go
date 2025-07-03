package logging

import (
	"strings"

	"go.uber.org/zap"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

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
