package sanitization_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	loggingsanitization "github.com/goformx/goforms/internal/infrastructure/logging/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

func TestUUIDSanitizationRule(t *testing.T) {
	sanitizer := sanitization.NewService()
	fieldSanitizer := loggingsanitization.NewFieldSanitizer()

	tests := []struct {
		name     string
		key      string
		value    any
		expected string
	}{
		{
			name:     "form_id should be masked",
			key:      "form_id",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e...0000",
		},
		{
			name:     "user_id should be masked",
			key:      "user_id",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e...0000",
		},
		{
			name:     "non-uuid field should not be masked",
			key:      "name",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "invalid uuid should be masked",
			key:      "user_id",
			value:    "invalid-uuid",
			expected: "inva...uuid",
		},
		{
			name:     "test_user_id should not be masked",
			key:      "test_user_id",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "request_id should not be masked",
			key:      "request_id",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fieldSanitizer.Sanitize(tt.key, tt.value, sanitizer)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultSanitizationRule(t *testing.T) {
	rule := &loggingsanitization.DefaultSanitizationRule{}
	sanitizer := sanitization.NewService()

	tests := []struct {
		name     string
		key      string
		value    any
		expected string
	}{
		{
			name:     "string value",
			key:      "message",
			value:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "integer value",
			key:      "count",
			value:    42,
			expected: "42",
		},
		{
			name:     "boolean value",
			key:      "enabled",
			value:    true,
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.Process(tt.key, tt.value, sanitizer)
			assert.Equal(t, tt.expected, result)
		})
	}
}
