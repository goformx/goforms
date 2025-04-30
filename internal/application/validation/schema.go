package validation

import "fmt"

// SchemaProvider defines a function type that returns a validation schema
type SchemaProvider func() any

// schemaRegistry holds all available validation schemas
var schemaRegistry = map[string]SchemaProvider{
	"signup": func() any {
		return map[string]any{
			"first_name": map[string]any{
				"type":    "string",
				"min":     1,
				"message": "First name is required",
			},
			"last_name": map[string]any{
				"type":    "string",
				"min":     1,
				"message": "Last name is required",
			},
			"email": map[string]any{
				"type":    "email",
				"message": "Please enter a valid email address",
			},
			"password": map[string]any{
				"type":    "password",
				"min":     8,
				"message": "Password must be at least 8 characters and contain uppercase, lowercase, number, and special character",
			},
			"confirm_password": map[string]any{
				"type":       "match",
				"matchField": "password",
				"message":    "Passwords do not match",
			},
		}
	},
	"login": func() any {
		return map[string]any{
			"email": map[string]any{
				"type":    "email",
				"message": "Please enter a valid email address",
			},
			"password": map[string]any{
				"type":    "required",
				"message": "Password is required",
			},
		}
	},
}

// GetSchema returns the validation schema for the given name
func GetSchema(schemaName string) (any, error) {
	provider, exists := schemaRegistry[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", schemaName)
	}
	return provider(), nil
}

// RegisterSchema adds a new schema to the registry
func RegisterSchema(name string, provider SchemaProvider) {
	schemaRegistry[name] = provider
}
