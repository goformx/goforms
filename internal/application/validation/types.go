package validation

// Rule represents a validation rule for a form field
type Rule struct {
	Type      string `json:"type"`
	Value     any    `json:"value,omitempty"`
	Message   string `json:"message,omitempty"`
	Condition string `json:"condition,omitempty"` // For conditional validation
}

// FieldValidation represents validation rules for a specific field
type FieldValidation struct {
	Required    bool           `json:"required,omitempty"`
	Type        string         `json:"type,omitempty"`
	MinLength   int            `json:"min_length,omitempty"`
	MaxLength   int            `json:"max_length,omitempty"`
	Min         float64        `json:"min,omitempty"`
	Max         float64        `json:"max,omitempty"`
	Pattern     string         `json:"pattern,omitempty"`
	Options     []string       `json:"options,omitempty"`
	CustomRules []Rule         `json:"custom_rules,omitempty"`
	Conditional map[string]any `json:"conditional,omitempty"`
}

// Error represents a validation error for a specific field
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Rule    string `json:"rule,omitempty"`
}

// Result represents the result of form validation
type Result struct {
	IsValid bool    `json:"is_valid"`
	Errors  []Error `json:"errors,omitempty"`
}

// FormValidatorInterface defines the interface for form validation
type FormValidatorInterface interface {
	ValidateForm(schema map[string]any, submission map[string]any) Result
	GenerateClientValidation(schema map[string]any) (map[string]any, error)
}

// FieldValidatorInterface defines the interface for field-specific validation
type FieldValidatorInterface interface {
	ValidateField(fieldName string, value any, rules FieldValidation) []Error
	ValidateFieldType(fieldName string, value any, fieldType string) *Error
}

// getMessage returns the default message for a validation rule.
// Custom per-ruleType message lookup is not implemented; callers may extend this later.
func (fv *FieldValidation) getMessage(ruleType, defaultMessage string) string {
	_ = ruleType // reserved for future per-rule message overrides
	return defaultMessage
}
