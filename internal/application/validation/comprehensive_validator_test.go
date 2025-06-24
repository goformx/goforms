package validation_test

import (
	"testing"

	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComprehensiveValidator_ValidateForm(t *testing.T) {
	validator := validation.NewComprehensiveValidator()

	tests := []struct {
		name           string
		schema         model.JSON
		submission     model.JSON
		expectedValid  bool
		expectedErrors int
		description    string
	}{
		{
			name: "valid submission with required fields",
			schema: model.JSON{
				"type": "object",
				"components": []any{
					map[string]any{
						"key":  "name",
						"type": "textfield",
						"validate": map[string]any{
							"required": true,
						},
					},
					map[string]any{
						"key":  "email",
						"type": "email",
						"validate": map[string]any{
							"required": true,
						},
					},
				},
			},
			submission: model.JSON{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			expectedValid:  true,
			expectedErrors: 0,
			description:    "Should validate submission with all required fields",
		},
		{
			name: "invalid submission missing required fields",
			schema: model.JSON{
				"type": "object",
				"components": []any{
					map[string]any{
						"key":  "name",
						"type": "textfield",
						"validate": map[string]any{
							"required": true,
						},
					},
					map[string]any{
						"key":  "email",
						"type": "email",
						"validate": map[string]any{
							"required": true,
						},
					},
				},
			},
			submission: model.JSON{
				"name": "John Doe",
				// email missing
			},
			expectedValid:  false,
			expectedErrors: 1,
			description:    "Should fail validation when required field is missing",
		},
		{
			name: "invalid schema missing components",
			schema: model.JSON{
				"type": "object",
				// components missing
			},
			submission: model.JSON{
				"name": "John Doe",
			},
			expectedValid:  false,
			expectedErrors: 1,
			description:    "Should fail validation when schema is invalid",
		},
		{
			name: "empty submission with optional fields",
			schema: model.JSON{
				"type": "object",
				"components": []any{
					map[string]any{
						"key":  "name",
						"type": "textfield",
						"validate": map[string]any{
							"required": false,
						},
					},
				},
			},
			submission:     model.JSON{},
			expectedValid:  true,
			expectedErrors: 0,
			description:    "Should validate empty submission with optional fields",
		},
		{
			name: "complex validation with multiple rules",
			schema: model.JSON{
				"type": "object",
				"components": []any{
					map[string]any{
						"key":  "age",
						"type": "number",
						"validate": map[string]any{
							"required": true,
							"min":      float64(18),
							"max":      float64(100),
						},
					},
					map[string]any{
						"key":  "password",
						"type": "password",
						"validate": map[string]any{
							"required":  true,
							"minLength": float64(8),
						},
					},
				},
			},
			submission: model.JSON{
				"age":      float64(25),
				"password": "securepass123",
			},
			expectedValid:  true,
			expectedErrors: 0,
			description:    "Should validate complex validation rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateForm(tt.schema, tt.submission)

			assert.Equal(t, tt.expectedValid, result.IsValid)
			assert.Len(t, result.Errors, tt.expectedErrors)

			if !tt.expectedValid && len(result.Errors) > 0 {
				// Check that error messages are meaningful
				for _, err := range result.Errors {
					assert.NotEmpty(t, err.Field)
					assert.NotEmpty(t, err.Message)
				}
			}
		})
	}
}

func TestComprehensiveValidator_GenerateClientValidation(t *testing.T) {
	validator := validation.NewComprehensiveValidator()

	tests := []struct {
		name           string
		schema         model.JSON
		expectedFields int
		description    string
	}{
		{
			name: "simple form with text fields",
			schema: model.JSON{
				"type": "object",
				"components": []any{
					map[string]any{
						"key":  "name",
						"type": "textfield",
						"validate": map[string]any{
							"required": true,
						},
					},
					map[string]any{
						"key":  "email",
						"type": "email",
						"validate": map[string]any{
							"required": true,
						},
					},
				},
			},
			expectedFields: 2,
			description:    "Should generate validation for text fields",
		},
		{
			name: "form with complex validation rules",
			schema: model.JSON{
				"type": "object",
				"components": []any{
					map[string]any{
						"key":  "age",
						"type": "number",
						"validate": map[string]any{
							"required": true,
							"min":      float64(18),
							"max":      float64(100),
						},
					},
					map[string]any{
						"key":  "password",
						"type": "password",
						"validate": map[string]any{
							"required":  true,
							"minLength": float64(8),
						},
					},
				},
			},
			expectedFields: 2,
			description:    "Should generate validation for complex rules",
		},
		{
			name: "form without components",
			schema: model.JSON{
				"type": "object",
			},
			expectedFields: 0,
			description:    "Should handle form without components",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientValidation, err := validator.GenerateClientValidation(tt.schema)

			if tt.name == "form without components" {
				// This should return an error for invalid schema
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid schema: missing components")
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, clientValidation)

			// Check that the generated validation has the expected structure
			// clientValidation is already map[string]any, no type assertion needed
			validationMap := clientValidation

			// Count the number of fields with validation rules
			fieldCount := 0
			for key, value := range validationMap {
				if key != "type" && key != "components" {
					fieldCount++
					// Check that each field has validation rules
					fieldRules, ok := value.(map[string]any)
					assert.True(t, ok, "Field validation should be a map")
					assert.NotEmpty(t, fieldRules, "Field should have validation rules")
				}
			}

			assert.Equal(t, tt.expectedFields, fieldCount)
		})
	}
}

func TestComprehensiveValidator_ValidateComponent(t *testing.T) {
	validator := validation.NewComprehensiveValidator()

	tests := []struct {
		name           string
		component      map[string]any
		submission     model.JSON
		expectedErrors int
		description    string
	}{
		{
			name: "required field present",
			component: map[string]any{
				"key":  "name",
				"type": "textfield",
				"validate": map[string]any{
					"required": true,
				},
			},
			submission: model.JSON{
				"name": "John Doe",
			},
			expectedErrors: 0,
			description:    "Should validate required field when present",
		},
		{
			name: "required field missing",
			component: map[string]any{
				"key":  "name",
				"type": "textfield",
				"validate": map[string]any{
					"required": true,
				},
			},
			submission: model.JSON{
				// name missing
			},
			expectedErrors: 1,
			description:    "Should fail validation when required field is missing",
		},
		{
			name: "required field empty string",
			component: map[string]any{
				"key":  "name",
				"type": "textfield",
				"validate": map[string]any{
					"required": true,
				},
			},
			submission: model.JSON{
				"name": "",
			},
			expectedErrors: 1,
			description:    "Should fail validation when required field is empty",
		},
		{
			name: "optional field missing",
			component: map[string]any{
				"key":  "description",
				"type": "textarea",
				"validate": map[string]any{
					"required": false,
				},
			},
			submission: model.JSON{
				// description missing
			},
			expectedErrors: 0,
			description:    "Should validate optional field when missing",
		},
		{
			name: "number field with min/max validation",
			component: map[string]any{
				"key":  "age",
				"type": "number",
				"validate": map[string]any{
					"required": true,
					"min":      float64(18),
					"max":      float64(100),
				},
			},
			submission: model.JSON{
				"age": float64(25),
			},
			expectedErrors: 0,
			description:    "Should validate number field within range",
		},
		{
			name: "number field below minimum",
			component: map[string]any{
				"key":  "age",
				"type": "number",
				"validate": map[string]any{
					"required": true,
					"min":      float64(18),
					"max":      float64(100),
				},
			},
			submission: model.JSON{
				"age": float64(16),
			},
			expectedErrors: 1,
			description:    "Should fail validation when number is below minimum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use reflection to access the private validateComponent method
			// Since it's private, we'll test it indirectly through ValidateForm
			schema := model.JSON{
				"type":       "object",
				"components": []any{tt.component},
			}

			result := validator.ValidateForm(schema, tt.submission)

			if tt.expectedErrors == 0 {
				assert.True(t, result.IsValid)
			} else {
				assert.False(t, result.IsValid)
				assert.Len(t, result.Errors, tt.expectedErrors)
			}
		})
	}
}

func TestComprehensiveValidator_ErrorMessages(t *testing.T) {
	validator := validation.NewComprehensiveValidator()

	schema := model.JSON{
		"type": "object",
		"components": []any{
			map[string]any{
				"key":  "name",
				"type": "textfield",
				"validate": map[string]any{
					"required": true,
				},
			},
			map[string]any{
				"key":  "email",
				"type": "email",
				"validate": map[string]any{
					"required": true,
				},
			},
		},
	}

	submission := model.JSON{
		// Both fields missing
	}

	result := validator.ValidateForm(schema, submission)

	assert.False(t, result.IsValid)
	assert.Len(t, result.Errors, 2)

	// Check that error messages are meaningful
	for _, err := range result.Errors {
		assert.NotEmpty(t, err.Field)
		assert.NotEmpty(t, err.Message)
		assert.Contains(t, err.Message, "required")
	}
}

func TestComprehensiveValidator_EdgeCases(t *testing.T) {
	validator := validation.NewComprehensiveValidator()

	tests := []struct {
		name        string
		schema      model.JSON
		submission  model.JSON
		shouldPanic bool
		description string
	}{
		{
			name:   "nil schema",
			schema: nil,
			submission: model.JSON{
				"name": "test",
			},
			shouldPanic: false,
			description: "Should handle nil schema gracefully",
		},
		{
			name: "nil submission",
			schema: model.JSON{
				"type":       "object",
				"components": []any{},
			},
			submission:  nil,
			shouldPanic: false,
			description: "Should handle nil submission gracefully",
		},
		{
			name:   "empty schema",
			schema: model.JSON{},
			submission: model.JSON{
				"name": "test",
			},
			shouldPanic: false,
			description: "Should handle empty schema gracefully",
		},
		{
			name: "schema without type",
			schema: model.JSON{
				"components": []any{},
			},
			submission: model.JSON{
				"name": "test",
			},
			shouldPanic: false,
			description: "Should handle schema without type gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					validator.ValidateForm(tt.schema, tt.submission)
				})
			} else {
				// Should not panic
				result := validator.ValidateForm(tt.schema, tt.submission)
				assert.NotNil(t, result)
			}
		})
	}
}
