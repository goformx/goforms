package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form/model"
)

func setupTestComprehensiveValidator() *validation.ComprehensiveValidator {
	validator := validation.NewComprehensiveValidator()

	return validator
}

func TestComprehensiveValidator_ValidateForm(t *testing.T) {
	validator := setupTestComprehensiveValidator()

	tests := []struct {
		name       string
		schema     model.JSON
		submission model.JSON
		wantValid  bool
	}{
		{
			name: "valid submission",
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
				},
			},
			submission: model.JSON{
				"name": "John Doe",
			},
			wantValid: true,
		},
		{
			name: "missing required field",
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
				},
			},
			submission: model.JSON{
				"email": "john@example.com",
			},
			wantValid: false,
		},
		{
			name: "invalid schema",
			schema: model.JSON{
				"invalid": "schema",
			},
			submission: model.JSON{
				"name": "John Doe",
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateForm(tt.schema, tt.submission)
			if tt.wantValid {
				require.True(t, result.IsValid)
			} else {
				require.False(t, result.IsValid)
			}
		})
	}
}

func TestComprehensiveValidator_GenerateClientValidation(t *testing.T) {
	validator := setupTestComprehensiveValidator()

	tests := []struct {
		name           string
		schema         model.JSON
		expectedFields []string
		description    string
	}{
		{
			name: "simple text field",
			schema: model.JSON{
				"type": "object",
				"components": []any{
					map[string]any{
						"key":  "name",
						"type": "textfield",
						"validate": map[string]any{
							"required":  true,
							"minLength": 3,
						},
					},
				},
			},
			expectedFields: []string{"name"},
			description:    "Should generate validation for text field",
		},
		{
			name: "multiple fields",
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
							"pattern":  "^[^@]+@[^@]+\\.[^@]+$",
						},
					},
				},
			},
			expectedFields: []string{"name", "email"},
			description:    "Should generate validation for multiple fields",
		},
		{
			name: "no validation rules",
			schema: model.JSON{
				"type": "object",
				"components": []any{
					map[string]any{
						"key":  "notes",
						"type": "textarea",
					},
				},
			},
			expectedFields: []string{},
			description:    "Should handle fields without validation rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientValidation, err := validator.GenerateClientValidation(tt.schema)
			require.NoError(t, err)

			// Check that validation object is created
			assert.NotNil(t, clientValidation)

			// Check that expected fields are present
			for _, field := range tt.expectedFields {
				assert.Contains(t, clientValidation, field)
			}
		})
	}
}

func TestComprehensiveValidator_EdgeCases(t *testing.T) {
	validator := setupTestComprehensiveValidator()

	tests := []struct {
		name       string
		schema     model.JSON
		submission model.JSON
		wantValid  bool
	}{
		{
			name:   "empty schema",
			schema: model.JSON{},
			submission: model.JSON{
				"name": "John Doe",
			},
			wantValid: false,
		},
		{
			name: "schema with no components",
			schema: model.JSON{
				"type":       "object",
				"components": []any{},
			},
			submission: model.JSON{
				"name": "John Doe",
			},
			wantValid: true, // Empty components array is considered valid
		},
		{
			name: "nil submission",
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
				},
			},
			submission: nil,
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateForm(tt.schema, tt.submission)
			if tt.wantValid {
				require.True(t, result.IsValid)
			} else {
				require.False(t, result.IsValid)
			}
		})
	}
}

func TestComprehensiveValidator_Integration(t *testing.T) {
	validator := setupTestComprehensiveValidator()

	// Create a complete form schema with multiple field types
	schema := model.JSON{
		"type": "object",
		"components": []any{
			map[string]any{
				"key":   "name",
				"type":  "textfield",
				"label": "Full Name",
				"validate": map[string]any{
					"required":  true,
					"minLength": 2,
				},
			},
			map[string]any{
				"key":   "email",
				"type":  "email",
				"label": "Email Address",
				"validate": map[string]any{
					"required": true,
					"pattern":  "^[^@]+@[^@]+\\.[^@]+$",
				},
			},
			map[string]any{
				"key":   "age",
				"type":  "number",
				"label": "Age",
				"validate": map[string]any{
					"min": 18,
					"max": 120,
				},
			},
		},
	}

	// Test client validation generation
	clientValidation, err := validator.GenerateClientValidation(schema)
	require.NoError(t, err)
	assert.NotNil(t, clientValidation)
	assert.Contains(t, clientValidation, "name")
	assert.Contains(t, clientValidation, "email")
	assert.Contains(t, clientValidation, "age")

	// Test valid submission
	validSubmission := model.JSON{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   25,
	}
	result := validator.ValidateForm(schema, validSubmission)
	require.True(t, result.IsValid)

	// Test invalid submission
	invalidSubmission := model.JSON{
		"name":  "J", // Too short
		"email": "invalid-email",
		"age":   15, // Too young
	}
	result = validator.ValidateForm(schema, invalidSubmission)
	require.False(t, result.IsValid)
	assert.NotEmpty(t, result.Errors)
}
