package model

import (
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/google/uuid"
)

// Form represents a form in the system
type Form struct {
	ID          string    `json:"id" db:"uuid"`
	UserID      uint      `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Schema      JSON      `json:"schema" db:"schema"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
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

	// Validate each field if there are any
	for i, field := range fields {
		fieldMap, fieldOk := field.(map[string]any)
		if !fieldOk {
			return fmt.Errorf("invalid field format")
		}

		if err := f.validateField(fieldMap); err != nil {
			return errors.Wrap(err, errors.ErrCodeValidation, fmt.Sprintf("invalid field at index %d", i))
		}
	}

	return nil
}

// validateField validates a single form field
func (f *Form) validateField(field map[string]any) error {
	// Validate required fields
	if err := f.validateRequiredFields(field); err != nil {
		return err
	}

	// Validate field type
	if err := f.validateFieldType(field); err != nil {
		return err
	}

	// Validate field options
	if err := f.validateFieldOptions(field); err != nil {
		return err
	}

	return nil
}

func (f *Form) validateRequiredFields(field map[string]any) error {
	requiredFields := []string{"type", "label"}
	for _, required := range requiredFields {
		if _, ok := field[required]; !ok {
			return fmt.Errorf("missing required field: %s", required)
		}
	}
	return nil
}

func (f *Form) validateFieldType(field map[string]any) error {
	fieldType, ok := field["type"].(string)
	if !ok {
		return fmt.Errorf("invalid field type")
	}

	switch fieldType {
	case "text", "textarea", "email", "password", "number", "date", "time", "datetime", "select", "radio", "checkbox", "file":
		return nil
	default:
		return fmt.Errorf("unsupported field type: %s", fieldType)
	}
}

func (f *Form) validateFieldOptions(field map[string]any) error {
	fieldType, _ := field["type"].(string)
	if fieldType != "select" && fieldType != "radio" && fieldType != "checkbox" {
		return nil
	}

	options, optionsOk := field["options"].([]any)
	if !optionsOk {
		return fmt.Errorf("invalid options format")
	}

	if len(options) == 0 {
		return fmt.Errorf("empty options for %s field", fieldType)
	}

	for _, option := range options {
		optionMap, optionOk := option.(map[string]any)
		if !optionOk {
			return fmt.Errorf("invalid option format")
		}

		if _, ok := optionMap["label"]; !ok {
			return fmt.Errorf("missing label in option")
		}

		if _, ok := optionMap["value"]; !ok {
			return fmt.Errorf("missing value in option")
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
