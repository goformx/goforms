package logging

import (
	"fmt"
	"strings"

	"github.com/mrz1836/go-sanitize"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// minMaskableLength is the minimum length required for partial masking
const minMaskableLength = 4

// Field represents a structured log field with type information
type Field struct {
	Key   string
	Value any
	Type  FieldType
}

// FieldType represents the type of a field
type FieldType int

const (
	StringFieldType FieldType = iota
	IntFieldType
	FloatFieldType
	BoolFieldType
	ErrorFieldType
	ObjectFieldType
	UUIDFieldType
	PathFieldType
	UserAgentFieldType
	SensitiveFieldType
	// InvalidPathMessage is the message returned for invalid paths
	InvalidPathMessage = "[invalid path]"
	// UUIDDashCount is the number of dashes in a standard UUID
	UUIDDashCount = 4
)

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value, Type: StringFieldType}
}

// Int creates an integer field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value, Type: IntFieldType}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value, Type: IntFieldType}
}

// Float creates a float field
func Float(key string, value float64) Field {
	return Field{Key: key, Value: value, Type: FloatFieldType}
}

// Bool creates a boolean field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value, Type: BoolFieldType}
}

// Error creates an error field
func Error(key string, err error) Field {
	return Field{Key: key, Value: err, Type: ErrorFieldType}
}

// UUID creates a UUID field with automatic masking
func UUID(key, value string) Field {
	return Field{Key: key, Value: value, Type: UUIDFieldType}
}

// Path creates a path field with validation
func Path(key, value string) Field {
	return Field{Key: key, Value: value, Type: PathFieldType}
}

// UserAgent creates a user agent field with sanitization
func UserAgent(key, value string) Field {
	return Field{Key: key, Value: value, Type: UserAgentFieldType}
}

// Sensitive creates a sensitive field that will be masked
func Sensitive(key string, value any) Field {
	return Field{Key: key, Value: value, Type: SensitiveFieldType}
}

// Object creates an object field for complex data
func Object(key string, obj any) Field {
	return Field{Key: key, Value: obj, Type: ObjectFieldType}
}

// ToZapField converts a Field to a zap.Field
func (f Field) ToZapField() zap.Field {
	// Check for sensitive fields first
	if isSensitiveKey(f.Key) {
		return zap.String(f.Key, "****")
	}

	return f.convertByType()
}

// convertByType converts the field based on its type
func (f Field) convertByType() zap.Field {
	converter, exists := fieldTypeConverters[f.Type]
	if !exists {
		return zap.Any(f.Key, f.Value)
	}

	return converter(f)
}

// fieldTypeConverters maps field types to their conversion functions
var fieldTypeConverters = map[FieldType]func(Field) zap.Field{
	StringFieldType: func(f Field) zap.Field {
		if str, ok := f.Value.(string); ok {
			return zap.String(f.Key, str)
		}

		return zap.String(f.Key, fmt.Sprintf("%v", f.Value))
	},
	IntFieldType: func(f Field) zap.Field {
		return f.convertIntField()
	},
	FloatFieldType: func(f Field) zap.Field {
		if val, ok := f.Value.(float64); ok {
			return zap.Float64(f.Key, val)
		}

		return zap.Any(f.Key, f.Value)
	},
	BoolFieldType: func(f Field) zap.Field {
		if val, ok := f.Value.(bool); ok {
			return zap.Bool(f.Key, val)
		}

		return zap.Any(f.Key, f.Value)
	},
	ErrorFieldType: func(f Field) zap.Field {
		return f.convertErrorField()
	},
	UUIDFieldType: func(f Field) zap.Field {
		if val, ok := f.Value.(string); ok {
			return zap.String(f.Key, maskUUID(val))
		}

		return zap.Any(f.Key, f.Value)
	},
	PathFieldType: func(f Field) zap.Field {
		if val, ok := f.Value.(string); ok {
			return zap.String(f.Key, sanitizePath(val))
		}

		return zap.Any(f.Key, f.Value)
	},
	UserAgentFieldType: func(f Field) zap.Field {
		if val, ok := f.Value.(string); ok {
			return zap.String(f.Key, sanitizeUserAgent(val))
		}

		return zap.Any(f.Key, f.Value)
	},
	ObjectFieldType: func(f Field) zap.Field {
		return zap.Any(f.Key, f.Value)
	},
	SensitiveFieldType: func(f Field) zap.Field {
		return zap.String(f.Key, "****")
	},
}

// convertIntField converts an integer field with type checking
func (f Field) convertIntField() zap.Field {
	switch v := f.Value.(type) {
	case int:
		return zap.Int(f.Key, v)
	case int64:
		return zap.Int64(f.Key, v)
	default:
		return zap.Any(f.Key, f.Value)
	}
}

// convertErrorField converts an error field with proper error handling
func (f Field) convertErrorField() zap.Field {
	if err, ok := f.Value.(error); ok {
		return zap.Error(err)
	}

	return zap.String(f.Key, fmt.Sprintf("%v", f.Value))
}

// maskUUID masks a UUID value for security
func maskUUID(value string) string {
	const uuidLength = 36

	const uuidDashCount = 4

	if len(value) == uuidLength && strings.Count(value, "-") == uuidDashCount {
		// Standard UUID format: mask middle part
		return value[:8] + "..." + value[len(value)-4:]
	}

	return value
}

// sanitizePath sanitizes a path value
func sanitizePath(value string) string {
	if value == "" || !strings.HasPrefix(value, "/") {
		return InvalidPathMessage
	}

	// Check for dangerous characters
	dangerousChars := []string{"\\", "<", ">", "\"", "'", "\x00", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return InvalidPathMessage
		}
	}

	// Check for path traversal attempts
	if strings.Contains(value, "..") || strings.Contains(value, "//") {
		return InvalidPathMessage
	}

	// Truncate if too long
	if len(value) > MaxPathLength {
		value = value[:MaxPathLength] + "..."
	}

	return sanitize.SingleLine(value)
}

// sanitizeUserAgent sanitizes a user agent value
func sanitizeUserAgent(value string) string {
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

// isSensitiveKey checks if a key contains sensitive data patterns
func isSensitiveKey(key string) bool {
	sensitivePatterns := []string{
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
		"oauth_refresh", "oauth_access", "oauth_id", "oauth_key",
		"data", "user_data", "personal_data", "sensitive_data",
	}

	keyLower := strings.ToLower(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(keyLower, pattern) {
			return true
		}
	}

	return false
}

// Legacy field constructors for backward compatibility
// These will be deprecated in favor of the new Field-based API

// SensitiveField creates a field that automatically masks sensitive data
func SensitiveField(key string, value any) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	return zap.Any(key, value)
}

// Sanitized creates a field with sanitized string data
func Sanitized(key, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	return zap.String(key, sanitize.SingleLine(value))
}

// SafeString creates a field with a safe string value (no sanitization)
func SafeString(key, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	return zap.String(key, value)
}

// RequestID creates a field with validated request ID
func RequestID(key, value string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	// Validate UUID format for request ID
	const uuidLength = 36

	const uuidDashCount = 4

	if len(value) == uuidLength && strings.Count(value, "-") == uuidDashCount {
		return zap.String(key, value)
	}

	return zap.String(key, "[invalid request id]")
}

// CustomField creates a field with custom sanitization logic
func CustomField(key string, value any, sanitizer func(any) string) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
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
	if len(value) <= minMaskableLength {
		// For short values, just return the mask
		return zap.String(key, mask)
	}

	// For longer values, show first 2 and last 2 characters with mask in middle
	maskedValue := value[:2] + mask + value[len(value)-2:]

	return zap.String(key, maskedValue)
}

// TruncatedField creates a field with truncated value
func TruncatedField(key, value string, maxLength int) zap.Field {
	if isSensitiveKey(key) {
		return zap.String(key, "****")
	}

	if len(value) > maxLength {
		value = value[:maxLength] + "..."
	}

	return zap.String(key, sanitize.SingleLine(value))
}

// ObjectField creates a field with sanitized object data
func ObjectField(key string, obj any) zap.Field {
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
	value any
}

// NewSensitiveObject creates a new sensitive object field
func NewSensitiveObject(key string, value any) SensitiveObject {
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
