// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// FormResponseHelper handles form-related HTTP responses
type FormResponseHelper struct{}

// NewFormResponseHelper creates a new FormResponseHelper
func NewFormResponseHelper() *FormResponseHelper {
	return &FormResponseHelper{}
}

// HandleCreateFormError handles errors from form creation
func (r *FormResponseHelper) HandleCreateFormError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrFormTitleRequired):
		return response.ErrorResponse(c, http.StatusBadRequest, "Form title is required")
	case errors.Is(err, model.ErrFormSchemaRequired):
		return response.ErrorResponse(c, http.StatusBadRequest, "Form schema is required")
	default:
		return response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create form")
	}
}

// SendCreateFormSuccess sends a successful form creation response
func (r *FormResponseHelper) SendCreateFormSuccess(c echo.Context, formID string) error {
	return response.Success(c, map[string]string{
		"message": "Form created successfully",
		"form_id": formID,
	})
}

// SendUpdateFormSuccess sends a successful form update response
func (r *FormResponseHelper) SendUpdateFormSuccess(c echo.Context, formID string) error {
	return response.Success(c, map[string]string{
		"message": "Form updated successfully",
		"form_id": formID,
	})
}

// SendDeleteFormSuccess sends a successful form deletion response
func (r *FormResponseHelper) SendDeleteFormSuccess(c echo.Context) error {
	return fmt.Errorf("send delete form success: %w", c.NoContent(constants.StatusNoContent))
}
