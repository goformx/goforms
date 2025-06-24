package validation

// ValidationRule represents a validation rule for a form field
type ValidationRule struct {
	Type      string `json:"type"`
	Value     any    `json:"value,omitempty"`
	Message   string `json:"message,omitempty"`
	Condition string `json:"condition,omitempty"` // For conditional validation
}

// FieldValidation represents validation rules for a specific field
type FieldValidation struct {
	Required    bool             `json:"required,omitempty"`
	Type        string           `json:"type,omitempty"`
	MinLength   int              `json:"minLength,omitempty"`
	MaxLength   int              `json:"maxLength,omitempty"`
	Min         float64          `json:"min,omitempty"`
	Max         float64          `json:"max,omitempty"`
	Pattern     string           `json:"pattern,omitempty"`
	Options     []string         `json:"options,omitempty"`
	CustomRules []ValidationRule `json:"customRules,omitempty"`
	Conditional map[string]any   `json:"conditional,omitempty"`
}

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Rule    string `json:"rule,omitempty"`
}

// ValidationResult represents the result of form validation
type ValidationResult struct {
	IsValid bool              `json:"isValid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// FormValidatorInterface defines the interface for form validation
type FormValidatorInterface interface {
	ValidateForm(schema map[string]any, submission map[string]any) ValidationResult
	GenerateClientValidation(schema map[string]any) (map[string]any, error)
}

// FieldValidatorInterface defines the interface for field-specific validation
type FieldValidatorInterface interface {
	ValidateField(fieldName string, value any, rules FieldValidation) []ValidationError
	ValidateFieldType(fieldName string, value any, fieldType string) *ValidationError
}

// getMessage returns a custom message or default message
func (fv FieldValidation) getMessage(ruleType, defaultMessage string) string {
	// TODO: Implement custom message lookup based on ruleType
	// For now, return the default message
	_ = ruleType // Suppress unused parameter warning
	return defaultMessage
}
