package model

import (
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/google/uuid"
)

// Form represents a form in the system
type Form struct {
	ID          string    `json:"id"`
	UserID      uint      `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Schema      JSON      `json:"schema"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JSON is a type alias for map[string]any to represent JSON data
type JSON map[string]any

// NewForm creates a new form instance
func NewForm(userID uint, title, description string, schema JSON) *Form {
	now := time.Now()
	return &Form{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Schema:      schema,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate validates the form
func (f *Form) Validate() error {
	if f.Title == "" {
		return ErrFormTitleRequired
	}

	if len(f.Title) < 3 {
		return errors.New(errors.ErrCodeValidation, "form title must be at least 3 characters long", nil)
	}

	if len(f.Title) > 100 {
		return errors.New(errors.ErrCodeValidation, "form title must not exceed 100 characters", nil)
	}

	if f.Description != "" && len(f.Description) > 500 {
		return errors.New(errors.ErrCodeValidation, "form description must not exceed 500 characters", nil)
	}

	if f.Schema == nil {
		return ErrFormSchemaRequired
	}

	if len(f.Schema) == 0 {
		return errors.New(errors.ErrCodeValidation, "form schema cannot be empty", nil)
	}

	// Validate schema structure
	if err := f.validateSchema(); err != nil {
		return errors.Wrap(err, errors.ErrCodeValidation, "invalid form schema")
	}

	return nil
}

// validateSchema validates the form schema structure
func (f *Form) validateSchema() error {
	// Check for required schema fields
	requiredFields := []string{"fields", "title", "description"}
	for _, field := range requiredFields {
		if _, exists := f.Schema[field]; !exists {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("schema missing required field: %s", field), nil)
		}
	}

	// Validate fields array
	fields, ok := f.Schema["fields"].([]any)
	if !ok {
		return errors.New(errors.ErrCodeValidation, "schema fields must be an array", nil)
	}

	if len(fields) == 0 {
		return errors.New(errors.ErrCodeValidation, "schema must contain at least one field", nil)
	}

	// Validate each field
	for i, field := range fields {
		fieldMap, ok := field.(map[string]any)
		if !ok {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("invalid field structure at index %d", i), nil)
		}

		if err := f.validateField(fieldMap); err != nil {
			return errors.Wrap(err, errors.ErrCodeValidation, fmt.Sprintf("invalid field at index %d", i))
		}
	}

	return nil
}

// validateField validates a single form field
func (f *Form) validateField(field map[string]any) error {
	// Check required field properties
	requiredProps := []string{"type", "name", "label"}
	for _, prop := range requiredProps {
		if _, exists := field[prop]; !exists {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("field missing required property: %s", prop), nil)
		}
	}

	// Validate field type
	fieldType, ok := field["type"].(string)
	if !ok {
		return errors.New(errors.ErrCodeValidation, "field type must be a string", nil)
	}

	validTypes := []string{"text", "textarea", "number", "email", "select", "checkbox", "radio", "date"}
	validType := false
	for _, t := range validTypes {
		if fieldType == t {
			validType = true
			break
		}
	}

	if !validType {
		return errors.New(errors.ErrCodeValidation, fmt.Sprintf("invalid field type: %s", fieldType), nil)
	}

	// Validate field name
	name, ok := field["name"].(string)
	if !ok {
		return errors.New(errors.ErrCodeValidation, "field name must be a string", nil)
	}

	if len(name) < 1 {
		return errors.New(errors.ErrCodeValidation, "field name cannot be empty", nil)
	}

	// Validate field label
	label, ok := field["label"].(string)
	if !ok {
		return errors.New(errors.ErrCodeValidation, "field label must be a string", nil)
	}

	if len(label) < 1 {
		return errors.New(errors.ErrCodeValidation, "field label cannot be empty", nil)
	}

	// Validate field-specific properties
	switch fieldType {
	case "select", "radio":
		if options, exists := field["options"]; exists {
			optionsArray, ok := options.([]any)
			if !ok {
				return errors.New(errors.ErrCodeValidation, "field options must be an array", nil)
			}
			if len(optionsArray) == 0 {
				return errors.New(errors.ErrCodeValidation, "field must have at least one option", nil)
			}
		} else {
			return errors.New(errors.ErrCodeValidation, "field must have options", nil)
		}
	}

	return nil
}

// Update updates the form with new values
func (f *Form) Update(title, description string, schema JSON) {
	f.Title = title
	f.Description = description
	if schema != nil {
		f.Schema = schema
	}
	f.UpdatedAt = time.Now()
}

// Deactivate marks the form as inactive
func (f *Form) Deactivate() {
	f.Active = false
	f.UpdatedAt = time.Now()
}

// Activate marks the form as active
func (f *Form) Activate() {
	f.Active = true
	f.UpdatedAt = time.Now()
}
