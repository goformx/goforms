package logging

import (
	"go.uber.org/zap"

	loggingsanitization "github.com/goformx/goforms/internal/infrastructure/logging/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
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
