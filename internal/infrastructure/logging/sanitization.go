package logging

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-sanitize"
)

// Sanitizer provides simplified field sanitization
type Sanitizer struct {
	// Cache for repeated sanitization operations
	cache map[string]string
}

// NewSanitizer creates a new sanitizer instance
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		cache: make(map[string]string),
	}
}

// SanitizeField sanitizes a field value based on its key and type
func (s *Sanitizer) SanitizeField(key string, value any) string {
	// Check for sensitive fields first
	if isSensitiveKey(key) {
		return "****"
	}

	// Handle different value types
	switch v := value.(type) {
	case string:
		return s.sanitizeString(key, v)
	case error:
		return s.sanitizeError(v)
	default:
		// For complex types, convert to string and sanitize
		return s.sanitizeString(key, fmt.Sprintf("%v", v))
	}
}

// sanitizeString sanitizes a string value based on the field key
func (s *Sanitizer) sanitizeString(key, value string) string {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", key, value)
	if cached, exists := s.cache[cacheKey]; exists {
		return cached
	}

	var sanitized string

	switch {
	case key == "path" || strings.Contains(key, "path"):
		sanitized = s.sanitizePath(value)
	case key == "user_agent" || strings.Contains(key, "user_agent"):
		sanitized = s.sanitizeUserAgent(value)
	case key == "uuid" || strings.Contains(key, "uuid") || strings.Contains(key, "id"):
		sanitized = s.sanitizeUUID(value)
	case key == "error" || key == "err" || strings.Contains(key, "error"):
		sanitized = s.sanitizeErrorString(value)
	default:
		sanitized = s.sanitizeGenericString(value)
	}

	// Cache the result
	s.cache[cacheKey] = sanitized
	return sanitized
}

// sanitizePath sanitizes a path value
func (s *Sanitizer) sanitizePath(value string) string {
	if value == "" || !strings.HasPrefix(value, "/") {
		return "[invalid path]"
	}

	// Check for dangerous characters
	dangerousChars := []string{"\\", "<", ">", "\"", "'", "\x00", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return "[invalid path]"
		}
	}

	// Check for path traversal attempts
	if strings.Contains(value, "..") || strings.Contains(value, "//") {
		return "[invalid path]"
	}

	// Truncate if too long
	if len(value) > MaxPathLength {
		value = value[:MaxPathLength] + "..."
	}

	return sanitize.SingleLine(value)
}

// sanitizeUserAgent sanitizes a user agent value
func (s *Sanitizer) sanitizeUserAgent(value string) string {
	if value == "" {
		return "[empty user agent]"
	}

	// Check for dangerous characters
	if strings.ContainsAny(value, "\n\r\x00") {
		return "[invalid user agent]"
	}

	// Truncate if too long
	if len(value) > MaxUserAgentLength {
		value = value[:MaxUserAgentLength] + "..."
	}

	return sanitize.SingleLine(value)
}

// sanitizeUUID sanitizes a UUID value
func (s *Sanitizer) sanitizeUUID(value string) string {
	if len(value) == 36 && strings.Count(value, "-") == 4 {
		// Standard UUID format: mask middle part
		return value[:8] + "..." + value[len(value)-4:]
	}
	return value
}

// sanitizeErrorString sanitizes an error message string
func (s *Sanitizer) sanitizeErrorString(value string) string {
	if value == "" {
		return ""
	}
	return sanitize.SingleLine(value)
}

// sanitizeError sanitizes an error value
func (s *Sanitizer) sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	return s.sanitizeErrorString(err.Error())
}

// sanitizeGenericString sanitizes a generic string value
func (s *Sanitizer) sanitizeGenericString(value string) string {
	// Check for dangerous characters
	if strings.ContainsAny(value, "\n\r\x00<>\"'\\") {
		return sanitize.SingleLine(value)
	}

	// Truncate if too long
	if len(value) > MaxStringLength {
		value = value[:MaxStringLength] + "..."
	}

	return value
}

// ClearCache clears the sanitization cache
func (s *Sanitizer) ClearCache() {
	s.cache = make(map[string]string)
}

// GetCacheSize returns the current cache size
func (s *Sanitizer) GetCacheSize() int {
	return len(s.cache)
}
