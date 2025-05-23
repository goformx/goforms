package services

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// FormData represents the structure for form creation and updates
type FormData struct {
	Title       string `json:"title" form:"title" validate:"required"`
	Description string `json:"description" form:"description" validate:"required"`
}

// FormOperations handles common form operations
type FormOperations struct {
	formService form.Service
	logger      logging.Logger
}

// NewFormOperations creates a new form operations service
func NewFormOperations(formService form.Service, logger logging.Logger) *FormOperations {
	return &FormOperations{
		formService: formService,
		logger:      logger,
	}
}

// ValidateAndBindFormData validates and binds form data from the request
func (o *FormOperations) ValidateAndBindFormData(c echo.Context) (*FormData, error) {
	var formData FormData
	if err := c.Bind(&formData); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if err := c.Validate(&formData); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, "Form validation failed")
	}

	return &formData, nil
}

// EnsureFormOwnership checks if the user owns the form
func (o *FormOperations) EnsureFormOwnership(
	c echo.Context,
	currentUser *user.User,
	formID string,
) (*form.Form, error) {
	if formID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	formObj, err := o.formService.GetForm(formID)
	if err != nil {
		o.logger.Error("Failed to get form", err)
		return nil, echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	if formObj.UserID != currentUser.ID {
		return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	return formObj, nil
}

// CreateDefaultSchema creates a default form schema
func (o *FormOperations) CreateDefaultSchema() form.JSON {
	return form.JSON{
		"display":    "form",
		"components": []any{},
	}
}

// UpdateFormDetails updates a form's basic details
func (o *FormOperations) UpdateFormDetails(formObj *form.Form, formData *FormData) {
	formObj.Title = formData.Title
	formObj.Description = formData.Description
}
