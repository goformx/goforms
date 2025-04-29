package validation

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Field   string
	Type    string
	Params  map[string]interface{}
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
				Field:   "username",
				Type:    "string",
				Params:  map[string]interface{}{"min": 3, "max": 50, "pattern": "^[a-zA-Z0-9_]+$"},
				Message: "Username must be 3-50 characters and can only contain letters, numbers, and underscores",
			},
			{
				Field:   "email",
				Type:    "email",
				Message: "Please enter a valid email address",
			},
			{
				Field:   "password",
				Type:    "password",
				Params:  map[string]interface{}{"min": 8},
				Message: "Password must be at least 8 characters and contain uppercase, lowercase, number, and special character",
			},
			{
				Field:   "confirmPassword",
				Type:    "match",
				Params:  map[string]interface{}{"field": "password"},
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
