package sanitization

import (
	"fmt"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// FieldSanitizer manages multiple sanitization rules
type FieldSanitizer struct {
	rules []SanitizationRule
}

// NewFieldSanitizer creates a new field sanitizer with default rules
func NewFieldSanitizer() *FieldSanitizer {
	return &FieldSanitizer{
		rules: []SanitizationRule{
			&PathSanitizationRule{},
			&UserAgentSanitizationRule{},
			&UUIDSanitizationRule{},
			&ErrorSanitizationRule{},
			&DefaultSanitizationRule{},
		},
	}
}

// Sanitize applies the appropriate sanitization rule based on the field key
func (fs *FieldSanitizer) Sanitize(key string, value any, sanitizer sanitization.ServiceInterface) string {
	for _, rule := range fs.rules {
		if rule.Matches(key) {
			return rule.Process(key, value, sanitizer)
		}
	}
	return fmt.Sprintf("%v", value) // Fallback
}
