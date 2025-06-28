package validation

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// ValidateRequired is the validation rule for required fields
	ValidateRequired = "required"
	// ValidatePassword is the validation rule for password fields
	ValidatePassword = "password"
)

// FormValidator provides form-specific validation utilities
type FormValidator struct {
	logger logging.Logger
}

// NewFormValidator creates a new form validator
func NewFormValidator(logger logging.Logger) *FormValidator {
	return &FormValidator{
		logger: logger,
	}
}

// ValidateFormID validates that a form ID parameter exists
func (fv *FormValidator) ValidateFormID(c echo.Context) (string, error) {
	formID := c.Param("id")
	if formID == "" {
		return "", errors.New("form ID is required")
	}

	return formID, nil
}

// ValidateFormData validates form data against a schema
func (fv *FormValidator) ValidateFormData(data, schema map[string]any) error {
	// Basic validation - check if required fields are present
	for fieldName, fieldSchema := range schema {
		if err := fv.validateField(fieldName, fieldSchema, data); err != nil {
			return err
		}
	}

	return nil
}

// validateField validates a single field against the schema
func (fv *FormValidator) validateField(fieldName string, fieldSchema any, data map[string]any) error {
	fieldSchemaMap, ok := fieldSchema.(map[string]any)
	if !ok {
		return nil // Skip non-map field schemas
	}

	validate, hasValidate := fieldSchemaMap["validate"].(string)
	if !hasValidate {
		return nil // Skip fields without validation rules
	}

	if validate == ValidateRequired {
		if value, exists := data[fieldName]; !exists || value == "" {
			return fmt.Errorf("field %s is required", fieldName)
		}
	}

	return nil
}

// ValidateFormSchema validates a form schema structure
func (fv *FormValidator) ValidateFormSchema(schema map[string]any) error {
	// Basic schema validation - check if schema has required structure
	if schema == nil {
		return errors.New("schema cannot be nil")
	}

	// Check if schema has basic form structure
	if _, hasType := schema["type"]; !hasType {
		return errors.New("schema must have a type field")
	}

	if _, hasComponents := schema["components"]; !hasComponents {
		return errors.New("schema must have a components field")
	}

	return nil
}
