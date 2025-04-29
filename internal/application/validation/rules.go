package validation

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Field   string
	Type    string
	Params  map[string]any
	Message string
}

// ValidationSchema represents a collection of validation rules for a form
type ValidationSchema struct {
	Name  string
	Rules []ValidationRule
}

// GetSignupSchema returns the validation rules for signup form
func GetSignupSchema() ValidationSchema {
	return ValidationSchema{
		Name: "signup",
		Rules: []ValidationRule{
			{
				Field:   "first_name",
				Type:    "string",
				Params:  map[string]any{"min": 1},
				Message: "First name is required",
			},
			{
				Field:   "last_name",
				Type:    "string",
				Params:  map[string]any{"min": 1},
				Message: "Last name is required",
			},
			{
				Field:   "email",
				Type:    "email",
				Message: "Please enter a valid email address",
			},
			{
				Field:   "password",
				Type:    "password",
				Params:  map[string]any{"min": 8},
				Message: "Password must be at least 8 characters and contain uppercase, lowercase, number, and special character",
			},
			{
				Field:   "confirm_password",
				Type:    "match",
				Params:  map[string]any{"field": "password"},
				Message: "Passwords do not match",
			},
		},
	}
}

// GetLoginSchema returns the validation rules for login form
func GetLoginSchema() ValidationSchema {
	return ValidationSchema{
		Name: "login",
		Rules: []ValidationRule{
			{
				Field:   "email",
				Type:    "email",
				Message: "Please enter a valid email address",
			},
			{
				Field:   "password",
				Type:    "required",
				Message: "Password is required",
			},
		},
	}
}
