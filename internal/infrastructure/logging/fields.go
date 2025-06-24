package logging

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-sanitize"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Sensitive data patterns for automatic detection
var sensitivePatterns = []string{
	"password", "token", "secret", "key", "credential", "authorization",
	"cookie", "session", "api_key", "access_token", "private_key",
	"public_key", "certificate", "ssn", "credit_card", "bank_account",
	"phone", "email", "address", "dob", "birth_date", "social_security",
	"tax_id", "driver_license", "passport", "national_id", "health_record",
	"medical_record", "insurance", "benefit", "salary", "compensation",
	"bank_routing", "bank_swift", "iban", "account_number", "pin",
	"cvv", "cvc", "security_code", "verification_code", "otp",
	"mfa_code", "2fa_code", "recovery_code", "backup_code", "reset_token",
	"activation_code", "verification_token", "invite_code", "referral_code",
	"promo_code", "discount_code", "coupon_code", "gift_card", "voucher",
	"license_key", "product_key", "serial_number", "activation_key",
	"registration_key", "subscription_key", "membership_key", "access_code",
	"security_key", "encryption_key", "decryption_key", "signing_key",
	"verification_key", "authentication_key", "session_key", "cookie_key",
	"csrf_token", "xsrf_token", "oauth_token", "oauth_secret", "oauth_verifier",
	"oauth_code", "oauth_state", "oauth_nonce", "oauth_scope", "oauth_grant",
	"oauth_refresh", "oauth_access", "oauth_id", "oauth_key", "form_id",
	"data", "user_data", "personal_data", "sensitive_data",
}

// isSensitiveKey checks if a key matches any sensitive pattern
func isSensitiveKey(key string) bool {
	keyLower := strings.ToLower(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(keyLower, pattern) {
			return true
		}
	}
	return false
}

// Sensitive creates a field that automatically masks sensitive data
func Sensitive(key string, value interface{}) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}
	return zap.Any(key, value)
}

// Sanitized creates a field with sanitized string data
func Sanitized(key string, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}
	return zap.String(key, sanitize.SingleLine(value))
}

// SafeString creates a field with a safe string value (no sanitization)
func SafeString(key string, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}
	return zap.String(key, value)
}

// UUID creates a field with masked UUID values
func UUID(key string, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	// Validate and mask UUID
	if len(value) == 36 && strings.Count(value, "-") == 4 {
		// Standard UUID format: mask middle part
		return zap.String(key, value[:8]+"..."+value[len(value)-4:])
	}

	return zap.String(key, value)
}

// Path creates a field with sanitized path data
func Path(key string, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
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
	if len(value) > 500 {
		value = value[:500] + "..."
	}

	return zap.String(key, sanitize.SingleLine(value))
}

// UserAgent creates a field with sanitized user agent data
func UserAgent(key string, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
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
	if len(value) > 1000 {
		value = value[:1000] + "..."
	}

	return zap.String(key, sanitize.SingleLine(value))
}

// Error creates a field with sanitized error data
func Error(key string, err error) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	if err == nil {
		return zap.String(key, "")
	}

	// Sanitize error message
	errMsg := sanitize.SingleLine(err.Error())
	return zap.String(key, errMsg)
}

// RequestID creates a field with validated request ID
func RequestID(key string, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	// Validate UUID format for request ID
	if len(value) == 36 && strings.Count(value, "-") == 4 {
		return zap.String(key, value)
	}

	return zap.String(key, "[invalid request id]")
}

// CustomField creates a field with custom sanitization logic
func CustomField(key string, value interface{}, sanitizer func(interface{}) string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	sanitizedValue := sanitizer(value)
	return zap.String(key, sanitizedValue)
}

// MaskedField creates a field with custom masking
func MaskedField(key string, value string, mask string) zap.Field {
	return zap.String(key, mask)
}

// TruncatedField creates a field with truncated value
func TruncatedField(key string, value string, maxLength int) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	if len(value) > maxLength {
		value = value[:maxLength] + "..."
	}

	return zap.String(key, sanitize.SingleLine(value))
}

// ObjectField creates a field with sanitized object data
func ObjectField(key string, obj interface{}) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	// Convert object to string and sanitize
	objStr := fmt.Sprintf("%v", obj)
	return zap.String(key, sanitize.SingleLine(objStr))
}

// SensitiveObject creates a custom field that implements zapcore.ObjectMarshaler
// for complex objects that need sensitive data masking
type SensitiveObject struct {
	key   string
	value interface{}
}

// NewSensitiveObject creates a new sensitive object field
func NewSensitiveObject(key string, value interface{}) SensitiveObject {
	return SensitiveObject{key: key, value: value}
}

// MarshalLogObject implements zapcore.ObjectMarshaler
func (s SensitiveObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if isSensitiveKey(s.key) {
		enc.AddString(s.key, "****")
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
