package logging

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-sanitize"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/goformx/goforms/internal/infrastructure/logging/sensitive"
)

// Sensitive creates a field that automatically masks sensitive data
func Sensitive(key string, value any) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}
	return zap.Any(key, value)
}

// Sanitized creates a field with sanitized string data
func Sanitized(key, value string) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}
	return zap.String(key, sanitize.SingleLine(value))
}

// SafeString creates a field with a safe string value (no sanitization)
func SafeString(key, value string) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}
	return zap.String(key, value)
}

// UUID creates a field with masked UUID values
func UUID(key, value string) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}

	// Validate and mask UUID
	if len(value) == 36 && strings.Count(value, "-") == 4 {
		// Standard UUID format: mask middle part
		return zap.String(key, value[:8]+"..."+value[len(value)-4:])
	}

	return zap.String(key, value)
}

// Path creates a field with sanitized path data
func Path(key, value string) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}

	// Basic path validation and sanitization
	if value == "" || !strings.HasPrefix(value, "/") {
		return zap.String(key, "[invalid path]")
	}

	// Check for dangerous characters
	dangerousChars := []string{"\\", "<", ">", "\"", "'", "\x00", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return zap.String(key, "[invalid path]")
		}
	}

	// Check for path traversal attempts
	if strings.Contains(value, "..") || strings.Contains(value, "//") {
		return zap.String(key, "[invalid path]")
	}

	// Truncate if too long
	if len(value) > MaxPathLength {
		value = value[:MaxPathLength] + "..."
	}

	return zap.String(key, sanitize.SingleLine(value))
}

// UserAgent creates a field with sanitized user agent data
func UserAgent(key, value string) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}

	// Basic user agent validation
	if value == "" {
		return zap.String(key, "[empty user agent]")
	}

	// Check for dangerous characters
	if strings.ContainsAny(value, "\n\r\x00") {
		return zap.String(key, "[invalid user agent]")
	}

	// Truncate if too long
	if len(value) > MaxUserAgentLength {
		value = value[:MaxUserAgentLength] + "..."
	}

	return zap.String(key, sanitize.SingleLine(value))
}

// Error creates a field with sanitized error data
func Error(key string, err error) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}

	if err == nil {
		return zap.String(key, "")
	}

	// Sanitize error message
	errMsg := sanitize.SingleLine(err.Error())
	return zap.String(key, errMsg)
}

// RequestID creates a field with validated request ID
func RequestID(key, value string) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}

	// Validate UUID format for request ID
	if len(value) == 36 && strings.Count(value, "-") == 4 {
		return zap.String(key, value)
	}

	return zap.String(key, "[invalid request id]")
}

// CustomField creates a field with custom sanitization logic
func CustomField(key string, value any, sanitizer func(any) string) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}

	sanitizedValue := sanitizer(value)
	return zap.String(key, sanitizedValue)
}

// MaskedField creates a field with custom masking applied to the value
func MaskedField(key, value, mask string) zap.Field {
	if value == "" {
		return zap.String(key, mask)
	}

	// Apply masking logic: show first and last characters with mask in middle
	if len(value) <= 4 {
		// For short values, just return the mask
		return zap.String(key, mask)
	}

	// For longer values, show first 2 and last 2 characters with mask in middle
	maskedValue := value[:2] + mask + value[len(value)-2:]
	return zap.String(key, maskedValue)
}

// TruncatedField creates a field with truncated value
func TruncatedField(key, value string, maxLength int) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}

	if len(value) > maxLength {
		value = value[:maxLength] + "..."
	}

	return zap.String(key, sanitize.SingleLine(value))
}

// ObjectField creates a field with sanitized object data
func ObjectField(key string, obj any) zap.Field {
	if sensitive.IsKey(key) {
		return zap.String(key, sensitive.MaskValue())
	}

	// Convert object to string and sanitize
	objStr := fmt.Sprintf("%v", obj)
	return zap.String(key, sanitize.SingleLine(objStr))
}

// SensitiveObject creates a custom field that implements zapcore.ObjectMarshaler
// for complex objects that need sensitive data masking
type SensitiveObject struct {
	key   string
	value any
}

// NewSensitiveObject creates a new sensitive object field
func NewSensitiveObject(key string, value any) SensitiveObject {
	return SensitiveObject{key: key, value: value}
}

// MarshalLogObject implements zapcore.ObjectMarshaler
func (s SensitiveObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if sensitive.IsKey(s.key) {
		enc.AddString(s.key, sensitive.MaskValue())
		return nil
	}

	// For non-sensitive objects, add as string
	objStr := fmt.Sprintf("%v", s.value)
	enc.AddString(s.key, sanitize.SingleLine(objStr))
	return nil
}

// Field returns the SensitiveObject as a zap.Field
func (s SensitiveObject) Field() zap.Field {
	return zap.Object(s.key, s)
}
