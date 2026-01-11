package model_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/domain/form/model"
)

func TestForm_Validate(t *testing.T) {
	tests := []struct {
		name        string
		form        *model.Form
		wantErr     bool
		errContains string
	}{
		{
			name: "valid form",
			form: model.NewForm(
				"user123",
				"Test Form",
				"A test form description",
				model.JSON{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
			),
			wantErr: false,
		},
		{
			name: "empty title",
			form: model.NewForm(
				"user123",
				"",
				"A test form description",
				model.JSON{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
			),
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name: "title too short",
			form: model.NewForm(
				"user123",
				"Te",
				"A test form description",
				model.JSON{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
			),
			wantErr:     true,
			errContains: "title must be between 3 and 100 characters",
		},
		{
			name: "description too long",
			form: model.NewForm(
				"user123",
				"Test Form",
				"A test form description that is way too long and exceeds the maximum length of 500 characters. "+
					"This is just a test to ensure that the validation works correctly. "+
					"We need to make sure that the description field is properly validated. "+
					"The description should not exceed 500 characters in length. "+
					"This is a very long description that should trigger the validation error. "+
					"We are adding more text to make sure we exceed the limit. "+
					"Just a few more characters to go... "+
					"Lorem ipsum dolor sit amet, consectetur adipiscing elit. "+
					"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. "+
					"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. "+
					"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. "+
					"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. "+
					"Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium.",
				model.JSON{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
			),
			wantErr:     true,
			errContains: "description must not exceed 500 characters",
		},
		{
			name: "invalid schema type",
			form: model.NewForm(
				"user123",
				"Test Form",
				"A test form description",
				model.JSON{
					"type": "array",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
			),
			wantErr:     true,
			errContains: "invalid schema: must have 'type: object' or 'display: form'",
		},
		{
			name: "valid Form.io format schema",
			form: model.NewForm(
				"user123",
				"Test Form",
				"A test form description",
				model.JSON{
					"display": "form",
					"components": []any{
						map[string]any{
							"type":  "textfield",
							"key":   "name",
							"label": "Name",
						},
						map[string]any{
							"type":  "button",
							"key":   "submit",
							"label": "Submit",
						},
					},
				},
			),
			wantErr: false,
		},
		{
			name: "missing schema properties",
			form: model.NewForm(
				"user123",
				"Test Form",
				"A test form description",
				model.JSON{
					"type": "object",
				},
			),
			wantErr:     true,
			errContains: "schema must contain either properties or components",
		},
		{
			name: "invalid property type",
			form: model.NewForm(
				"user123",
				"Test Form",
				"A test form description",
				model.JSON{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "invalid",
						},
					},
				},
			),
			wantErr:     true,
			errContains: "invalid type 'invalid' for property 'name'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.form.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
