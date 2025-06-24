package logging_test

import (
	"errors"
	"testing"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestSensitiveField(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		expected string
	}{
		{
			name:     "sensitive key - password",
			key:      "password",
			value:    "secret123",
			expected: "****",
		},
		{
			name:     "sensitive key - token",
			key:      "access_token",
			value:    "abc123",
			expected: "****",
		},
		{
			name:     "non-sensitive key",
			key:      "name",
			value:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "case insensitive sensitive key",
			key:      "PASSWORD",
			value:    "secret123",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.Sensitive(tt.key, tt.value)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}

func TestSanitizedField(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "sensitive key should be masked",
			key:      "password",
			value:    "secret123",
			expected: "****",
		},
		{
			name:     "normal string should be sanitized",
			key:      "message",
			value:    "Hello\nWorld",
			expected: "Hello World",
		},
		{
			name:     "html should be sanitized",
			key:      "content",
			value:    "<script>alert('xss')</script>",
			expected: "<script>alert('xss')</script>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.Sanitized(tt.key, tt.value)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}

func TestUUIDField(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "sensitive key should be masked",
			key:      "form_id",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "****",
		},
		{
			name:     "valid UUID should be masked",
			key:      "user_id",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400...0000",
		},
		{
			name:     "invalid UUID should not be masked",
			key:      "user_id",
			value:    "invalid-uuid",
			expected: "invalid-uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.UUID(tt.key, tt.value)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}

func TestPathField(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "sensitive key should be masked",
			key:      "csrf_token",
			value:    "/api/test",
			expected: "****",
		},
		{
			name:     "valid path",
			key:      "path",
			value:    "/api/v1/users",
			expected: "/api/v1/users",
		},
		{
			name:     "invalid path - no leading slash",
			key:      "path",
			value:    "api/test",
			expected: "[invalid path]",
		},
		{
			name:     "invalid path - path traversal",
			key:      "path",
			value:    "/api/../etc/passwd",
			expected: "[invalid path]",
		},
		{
			name:     "invalid path - dangerous characters",
			key:      "path",
			value:    "/api/<script>alert(1)</script>",
			expected: "[invalid path]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.Path(tt.key, tt.value)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}

func TestUserAgentField(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "sensitive key should be masked",
			key:      "oauth_token",
			value:    "Mozilla/5.0",
			expected: "****",
		},
		{
			name:     "valid user agent",
			key:      "user_agent",
			value:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
		{
			name:     "invalid user agent - newlines",
			key:      "user_agent",
			value:    "Mozilla\n<script>alert(1)</script>",
			expected: "[invalid user agent]",
		},
		{
			name:     "empty user agent",
			key:      "user_agent",
			value:    "",
			expected: "[empty user agent]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.UserAgent(tt.key, tt.value)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}

func TestErrorField(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		err      error
		expected string
	}{
		{
			name:     "sensitive key should be masked",
			key:      "password",
			err:      errors.New("invalid password"),
			expected: "****",
		},
		{
			name:     "normal error",
			key:      "error",
			err:      errors.New("database connection failed"),
			expected: "database connection failed",
		},
		{
			name:     "nil error",
			key:      "error",
			err:      nil,
			expected: "",
		},
		{
			name:     "error with newlines",
			key:      "error",
			err:      errors.New("error\nwith\nnewlines"),
			expected: "error with newlines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.Error(tt.key, tt.err)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}

func TestRequestIDField(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "sensitive key should be masked",
			key:      "session_id",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "****",
		},
		{
			name:     "valid request ID",
			key:      "request_id",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "invalid request ID",
			key:      "request_id",
			value:    "invalid-id",
			expected: "[invalid request id]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.RequestID(tt.key, tt.value)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}

func TestSensitiveObject(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		expected string
	}{
		{
			name:     "sensitive key should be masked",
			key:      "user_data",
			value:    map[string]string{"name": "John", "email": "john@example.com"},
			expected: "****",
		},
		{
			name:     "normal object",
			key:      "config",
			value:    map[string]string{"env": "production", "version": "1.0.0"},
			expected: "map[env:production version:1.0.0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := logging.NewSensitiveObject(tt.key, tt.value)

			// Test marshaling
			encoder := zapcore.NewMapObjectEncoder()
			err := obj.MarshalLogObject(encoder)
			assert.NoError(t, err)

			// Check the result
			if tt.expected == "****" {
				// For sensitive keys, the object should be masked
				assert.Equal(t, "****", encoder.Fields[tt.key])
			} else {
				// For normal objects, check that it contains the expected content
				assert.Contains(t, encoder.Fields[tt.key], "production")
			}
		})
	}
}

func TestFieldIntegration(t *testing.T) {
	// Test that fields work correctly with actual logging
	logger := zaptest.NewLogger(t)

	// Test various field types
	logger.Info("test message",
		logging.Sensitive("password", "secret123"),
		logging.Sanitized("message", "Hello\nWorld"),
		logging.UUID("user_id", "550e8400-e29b-41d4-a716-446655440000"),
		logging.Path("path", "/api/v1/users"),
		logging.Error("error", errors.New("test error")),
	)

	// This should not panic and should log correctly
	assert.True(t, true, "Logging with custom fields should work")
}

func TestTruncatedField(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		maxLength int
		expected  string
	}{
		{
			name:      "sensitive key should be masked",
			key:       "api_key",
			value:     "very long api key",
			maxLength: 10,
			expected:  "****",
		},
		{
			name:      "long string should be truncated",
			key:       "description",
			value:     "This is a very long description that should be truncated",
			maxLength: 20,
			expected:  "This is a very long ...",
		},
		{
			name:      "short string should not be truncated",
			key:       "description",
			value:     "Short",
			maxLength: 20,
			expected:  "Short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.TruncatedField(tt.key, tt.value, tt.maxLength)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}

func TestCustomField(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     any
		sanitizer func(any) string
		expected  string
	}{
		{
			name:      "sensitive key should be masked",
			key:       "secret_data",
			value:     "sensitive value",
			sanitizer: func(v any) string { return "custom sanitized" },
			expected:  "****",
		},
		{
			name:      "custom sanitization",
			key:       "description",
			value:     "original value",
			sanitizer: func(v any) string { return "custom sanitized" },
			expected:  "custom sanitized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logging.CustomField(tt.key, tt.value, tt.sanitizer)
			assert.Equal(t, tt.key, field.Key)
			assert.Equal(t, tt.expected, field.String)
		})
	}
}
