// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// FormResponseHelper handles form-related HTTP responses
type FormResponseHelper struct{}

// NewFormResponseHelper creates a new FormResponseHelper
func NewFormResponseHelper() *FormResponseHelper {
	return &FormResponseHelper{}
}

// FormSuccessResponse represents a successful form operation response
type FormSuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	FormID  string `json:"form_id,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
}

// HandleCreateFormError handles errors from form creation
func (r *FormResponseHelper) HandleCreateFormError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrFormTitleRequired):
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Message: "Form title is required",
		})
	case errors.Is(err, model.ErrFormSchemaRequired):
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Message: "Form schema is required",
		})
	default:
		return c.JSON(http.StatusInternalServerError, &ErrorResponse{
			Message: "Failed to create form",
		})
	}
}

// SendCreateFormSuccess sends a successful form creation response
func (r *FormResponseHelper) SendCreateFormSuccess(c echo.Context, formID string) error {
	return c.JSON(http.StatusOK, &FormSuccessResponse{
		Success: true,
		Message: "Form created successfully",
		FormID:  formID,
	})
}

// SendUpdateFormSuccess sends a successful form update response
func (r *FormResponseHelper) SendUpdateFormSuccess(c echo.Context, formID string) error {
	return c.JSON(http.StatusOK, &FormSuccessResponse{
		Success: true,
		Message: "Form updated successfully",
		FormID:  formID,
	})
}

// SendDeleteFormSuccess sends a successful form deletion response
func (r *FormResponseHelper) SendDeleteFormSuccess(c echo.Context) error {
	return c.NoContent(constants.StatusNoContent)
}
