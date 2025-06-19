package validation

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

const (
	// Validation rules
	ValidateRequired = "required"
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
		return "", errors.New("Form ID is required")
	}
	return formID, nil
}

// ValidateUserOwnership verifies that a resource belongs to the authenticated user
func (fv *FormValidator) ValidateUserOwnership(c echo.Context, resourceUserID, requestUserID string) error {
	if resourceUserID != requestUserID {
		fv.logger.Error("ownership verification failed",
			"resource_user_id", resourceUserID,
			"request_user_id", requestUserID)
		return response.WebErrorResponse(c, nil, http.StatusForbidden,
			"You don't have permission to access this resource")
	}

	return nil
}

// HandleFormValidationError handles form validation errors
func (fv *FormValidator) HandleFormValidationError(c echo.Context, message string) error {
	return response.WebErrorResponse(c, nil, http.StatusBadRequest, message)
}

// HandleFormError handles form-specific errors
func (fv *FormValidator) HandleFormError(c echo.Context, err error, message string) error {
	fv.logger.Error("form operation failed", "error", err, "message", message)
	return response.WebErrorResponse(c, nil, http.StatusInternalServerError, message)
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
