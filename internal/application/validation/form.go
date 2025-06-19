package validation

import (
	"errors"
	"net/http"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
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
func (fv *FormValidator) ValidateFormData(data map[string]any, schema map[string]any) error {
	// This is a placeholder for form data validation
	// In a real implementation, this would validate the data against the schema
	return nil
}

// ValidateFormSchema validates a form schema structure
func (fv *FormValidator) ValidateFormSchema(schema map[string]any) error {
	// This is a placeholder for schema validation
	// In a real implementation, this would validate the schema structure
	return nil
}
