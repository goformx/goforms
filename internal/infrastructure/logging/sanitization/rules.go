// Package sanitization provides utilities for sanitizing log fields and sensitive data in logs.
package sanitization

import (
	"fmt"
	"strings"

	"github.com/goformx/goforms/internal/infrastructure/logging/sensitive"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// Rule defines how to process different field types
type Rule interface {
	Matches(key string) bool
	Process(key string, value any, sanitizer sanitization.ServiceInterface) string
}

// PathSanitizationRule handles path field validation and sanitization
type PathSanitizationRule struct{}

// Matches checks if this rule applies to the given key
func (r *PathSanitizationRule) Matches(key string) bool {
	return key == "path"
}

// Process sanitizes path field values
func (r *PathSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	if sensitive.IsKey(key) {
		return sensitive.MaskValue()
	}
	processor := &FieldProcessor{
		validator:   adaptPathValidator,
		sanitizer:   adaptStringSanitizer,
		invalidMsg:  "[invalid path]",
		invalidType: "[invalid path type]",
		maxLength:   MaxPathLength,
	}
	return processor.Process(value, sanitizer)
}

// UserAgentSanitizationRule handles user agent field validation and sanitization
type UserAgentSanitizationRule struct{}

// Matches checks if this rule applies to the given key
func (r *UserAgentSanitizationRule) Matches(key string) bool {
	return key == "user_agent"
}

// Process sanitizes user agent field values
func (r *UserAgentSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	if sensitive.IsKey(key) {
		return sensitive.MaskValue()
	}
	processor := &FieldProcessor{
		validator:   adaptUserAgentValidator,
		sanitizer:   adaptStringSanitizer,
		invalidMsg:  "[invalid user agent]",
		invalidType: "[invalid user agent type]",
		maxLength:   MaxStringLength,
	}
	return processor.Process(value, sanitizer)
}

// isUUIDField checks if a field key represents a UUID field that should be masked
func isUUIDField(key string) bool {
	keyLower := strings.ToLower(key)
	if strings.Contains(keyLower, "test") {
		return false
	}
	if keyLower == "user_id" || keyLower == "form_id" || keyLower == "id" {
		return true
	}
	if strings.HasSuffix(keyLower, "_id") && keyLower != "request_id" && keyLower != "session_id" {
		return true
	}
	return false
}

// UUIDSanitizationRule handles UUID-like field validation and masking
type UUIDSanitizationRule struct{}

// Matches checks if this rule applies to the given key
func (r *UUIDSanitizationRule) Matches(key string) bool {
	return isUUIDField(key)
}

// Process sanitizes UUID field values
func (r *UUIDSanitizationRule) Process(key string, value any, _ sanitization.ServiceInterface) string {
	if sensitive.IsKey(key) {
		return sensitive.MaskValue()
	}
	if id, ok := value.(string); ok {
		if len(id) >= UUIDMinMaskLen {
			return id[:UUIDMaskPrefixLen] + "..." + id[len(id)-UUIDMaskSuffixLen:]
		}
		return fmt.Sprintf("[id:%d]", len(id))
	}
	return "[invalid uuid type]"
}

// ErrorSanitizationRule handles error field sanitization
type ErrorSanitizationRule struct{}

// Matches checks if this rule applies to the given key
func (r *ErrorSanitizationRule) Matches(key string) bool {
	return key == "error"
}

// Process sanitizes error field values
func (r *ErrorSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	if sensitive.IsKey(key) {
		return sensitive.MaskValue()
	}
	if err, ok := value.(error); ok {
		return sanitizeError(err, sanitizer)
	}
	return fmt.Sprintf("%v", value)
}

// DefaultSanitizationRule handles all other field types
type DefaultSanitizationRule struct{}

// Matches checks if this rule applies to the given key
func (r *DefaultSanitizationRule) Matches(_ string) bool {
	return true // Matches everything (should be last in the chain)
}

// Process sanitizes default field values
func (r *DefaultSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	if sensitive.IsKey(key) {
		return sensitive.MaskValue()
	}
	if str, ok := value.(string); ok {
		if sanitizer != nil {
			return sanitizer.SanitizeForLogging(truncateString(str, MaxStringLength))
		}
		return truncateString(str, MaxStringLength)
	}
	objStr := fmt.Sprintf("%v", value)
	if sanitizer != nil {
		return sanitizer.SanitizeForLogging(objStr)
	}
	return objStr
}
