package logging

import (
	"strings"

	loggingsanitization "github.com/goformx/goforms/internal/infrastructure/logging/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"go.uber.org/zap"
)

// convertToZapFields converts a slice of fields to zap fields
func convertToZapFields(fields []any, fieldSanitizer *loggingsanitization.FieldSanitizer) []zap.Field {
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
		// Use the field sanitizer for consistent processing
		// Note: We need to pass a sanitizer here, but the rules will handle their own sanitization
		sanitizedValue := fieldSanitizer.Sanitize(key, value, nil)

		// Always append as string since sanitization returns string
		zapFields = append(zapFields, zap.String(key, sanitizedValue))
	}
	return zapFields
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

// sanitizeString sanitizes a string for safe logging
func sanitizeString(s string, sanitizer sanitization.ServiceInterface) string {
	if sanitizer != nil {
		return sanitizer.SanitizeForLogging(s)
	}
	return s
}

// truncateString truncates a string to the maximum allowed length
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// isUUIDField checks if a field key represents a UUID field that should be masked
func isUUIDField(key string) bool {
	return strings.Contains(strings.ToLower(key), "id") &&
		!strings.Contains(strings.ToLower(key), "length") &&
		key != "request_id"
}
