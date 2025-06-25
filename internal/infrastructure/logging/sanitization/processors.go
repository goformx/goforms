// Package sanitization provides utilities for sanitizing log fields and sensitive data in logs.
package sanitization

import (
	"fmt"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// FieldProcessor handles validation and sanitization of specific field types
type FieldProcessor struct {
	validator   func(any) bool
	sanitizer   func(any, sanitization.ServiceInterface) string
	invalidMsg  string
	invalidType string
	maxLength   int
}

// Process applies validation and sanitization to a field value
func (fp *FieldProcessor) Process(value any, sanitizer sanitization.ServiceInterface) string {
	if str, ok := value.(string); ok {
		if !fp.validator(str) {
			return fp.invalidMsg
		}
		return fp.sanitizer(truncateString(str, fp.maxLength), sanitizer)
	}
	return fp.invalidType
}

// Helper functions to adapt existing validation and sanitization functions
func adaptPathValidator(path any) bool {
	if str, ok := path.(string); ok {
		return validatePath(str)
	}
	return false
}

func adaptUserAgentValidator(userAgent any) bool {
	if str, ok := userAgent.(string); ok {
		return validateUserAgent(str)
	}
	return false
}

func adaptStringSanitizer(value any, sanitizer sanitization.ServiceInterface) string {
	if str, ok := value.(string); ok {
		return sanitizeString(str, sanitizer)
	}
	return fmt.Sprintf("%v", value)
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
