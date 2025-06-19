package validation

import (
	"reflect"
)

// SchemaGenerator provides functionality to generate validation schemas from struct tags
type SchemaGenerator struct{}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator() *SchemaGenerator {
	return &SchemaGenerator{}
}

// getFieldSchema extracts validation schema from struct field tags
func (sg *SchemaGenerator) getFieldSchema(field reflect.StructField) map[string]any {
	fieldSchema := make(map[string]any)

	// Get validation tags
	validate := field.Tag.Get("validate")
	if validate != "" {
		fieldSchema["validate"] = validate
	}

	// Get min/max length
	minLen := field.Tag.Get("minlen")
	if minLen != "" {
		fieldSchema["minLength"] = minLen
	}
	maxLen := field.Tag.Get("maxlen")
	if maxLen != "" {
		fieldSchema["maxLength"] = maxLen
	}

	// Set type and message based on validation rules
	if validate != "" {
		if validate == "required,email" {
			fieldSchema["type"] = "email"
			fieldSchema["message"] = "Please enter a valid email address"
		} else if validate == "required" {
			fieldSchema["type"] = "string"
			fieldSchema["message"] = "This field is required"
		} else if validate == "required,min=8" {
			fieldSchema["type"] = "password"
			fieldSchema["min"] = "8"
			fieldSchema["message"] = "Password must be at least 8 characters long and include uppercase, lowercase, number, and special characters"
		} else if validate == "required,eqfield=password" {
			fieldSchema["type"] = "match"
			fieldSchema["matchField"] = "password"
			fieldSchema["message"] = "Passwords don't match"
		}
	}

	return fieldSchema
}

// GenerateValidationSchema generates a validation schema from a struct
func (sg *SchemaGenerator) GenerateValidationSchema(s any) map[string]any {
	t := reflect.TypeOf(s)
	schema := make(map[string]any)

	for i := range t.NumField() {
		field := t.Field(i)
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = field.Name
		}

		fieldSchema := sg.getFieldSchema(field)
		schema[fieldName] = fieldSchema
	}

	return schema
}

// GenerateLoginSchema generates the validation schema for login forms
func (sg *SchemaGenerator) GenerateLoginSchema() map[string]any {
	// Import the user package to get the Login struct
	// This is a placeholder - the actual struct should be imported
	type Login struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	return sg.GenerateValidationSchema(Login{})
}

// GenerateSignupSchema generates the validation schema for signup forms
func (sg *SchemaGenerator) GenerateSignupSchema() map[string]any {
	// Import the user package to get the Signup struct
	// This is a placeholder - the actual struct should be imported
	type Signup struct {
		Email           string `json:"email" validate:"required,email"`
		Password        string `json:"password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=password"`
		FirstName       string `json:"first_name" validate:"required"`
		LastName        string `json:"last_name" validate:"required"`
	}

	return sg.GenerateValidationSchema(Signup{})
}
