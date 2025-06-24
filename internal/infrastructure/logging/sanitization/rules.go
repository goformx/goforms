package sanitization

import (
	"fmt"
	"strings"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// SanitizationRule defines how to process different field types
type SanitizationRule interface {
	Matches(key string) bool
	Process(key string, value any, sanitizer sanitization.ServiceInterface) string
}

// PathSanitizationRule handles path field validation and sanitization
type PathSanitizationRule struct{}

func (r *PathSanitizationRule) Matches(key string) bool {
	return key == "path"
}

func (r *PathSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	// Check if this is a sensitive key first
	if isSensitiveKey(key) {
		return "****"
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

func (r *UserAgentSanitizationRule) Matches(key string) bool {
	return key == "user_agent"
}

func (r *UserAgentSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	// Check if this is a sensitive key first
	if isSensitiveKey(key) {
		return "****"
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

// UUIDSanitizationRule handles UUID-like field validation and masking
type UUIDSanitizationRule struct{}

func (r *UUIDSanitizationRule) Matches(key string) bool {
	return isUUIDField(key)
}

func (r *UUIDSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	// Check if this is a sensitive key first
	if isSensitiveKey(key) {
		return "****"
	}

	if id, ok := value.(string); ok {
		if !validateUUID(id) {
			return "[invalid uuid format]"
		}
		// For UUIDs, we return a masked version for security
		if len(id) >= UUIDMinMaskLen {
			return id[:UUIDMaskPrefixLen] + "..." + id[len(id)-UUIDMaskSuffixLen:]
		}
		return "[invalid uuid length]"
	}
	return "[invalid uuid type]"
}

// ErrorSanitizationRule handles error field sanitization
type ErrorSanitizationRule struct{}

func (r *ErrorSanitizationRule) Matches(key string) bool {
	return key == "error"
}

func (r *ErrorSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	// Check if this is a sensitive key first
	if isSensitiveKey(key) {
		return "****"
	}

	if err, ok := value.(error); ok {
		return sanitizeError(err, sanitizer)
	}
	return fmt.Sprintf("%v", value)
}

// DefaultSanitizationRule handles all other field types
type DefaultSanitizationRule struct{}

func (r *DefaultSanitizationRule) Matches(key string) bool {
	return true // Matches everything (should be last in the chain)
}

func (r *DefaultSanitizationRule) Process(key string, value any, sanitizer sanitization.ServiceInterface) string {
	// Check if this is a sensitive key first
	if isSensitiveKey(key) {
		return "****"
	}

	if str, ok := value.(string); ok {
		// Apply sanitization if sanitizer is available
		if sanitizer != nil {
			return sanitizer.SanitizeForLogging(truncateString(str, MaxStringLength))
		}
		return truncateString(str, MaxStringLength)
	}

	// For other types, convert to string and sanitize
	objStr := fmt.Sprintf("%v", value)
	if sanitizer != nil {
		return sanitizer.SanitizeForLogging(objStr)
	}
	return objStr
}

// isSensitiveKey checks if a key matches any sensitive pattern
func isSensitiveKey(key string) bool {
	sensitivePatterns := []string{
		"password", "passwd", "pwd", "secret", "key", "token", "auth", "credential",
		"private", "sensitive", "confidential", "hidden", "masked", "encrypted",
		"hash", "salt", "nonce", "challenge", "response", "signature", "certificate",
		"session", "cookie", "jwt", "bearer", "api_key", "apikey", "access_key",
		"secret_key", "private_key", "public_key", "ssh_key", "gpg_key", "pgp_key",
		"oauth_token", "oauth_secret", "oauth_key", "oauth_code", "oauth_state",
		"oauth_nonce", "oauth_scope", "oauth_grant", "oauth_refresh", "oauth_access",
		"oauth_id", "oauth_key", "form_id",
	}

	keyLower := strings.ToLower(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(keyLower, pattern) {
			return true
		}
	}
	return false
}

// isUUIDField checks if a field key represents a UUID field that should be masked
func isUUIDField(key string) bool {
	return strings.Contains(strings.ToLower(key), "id") &&
		!strings.Contains(strings.ToLower(key), "length") &&
		key != "request_id"
}
