package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// FormErrorHandlerImpl handles form-specific error scenarios
type FormErrorHandlerImpl struct {
	responseBuilder FormResponseBuilder
}

// NewFormErrorHandler creates a new form error handler
func NewFormErrorHandler(responseBuilder FormResponseBuilder) FormErrorHandler {
	return &FormErrorHandlerImpl{
		responseBuilder: responseBuilder,
	}
}

// HandleSchemaError handles schema-related errors
func (h *FormErrorHandlerImpl) HandleSchemaError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrFormSchemaRequired):
		return h.responseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form schema is required")
	case errors.Is(err, model.ErrFormInvalid):
		return h.responseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Invalid form schema format")
	default:
		return h.responseBuilder.BuildErrorResponse(c, http.StatusInternalServerError, "Failed to process form schema")
	}
}

// HandleSubmissionError handles form submission errors
func (h *FormErrorHandlerImpl) HandleSubmissionError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrFormNotFound):
		return h.responseBuilder.BuildErrorResponse(c, http.StatusNotFound, "Form not found")
	case errors.Is(err, model.ErrFormInvalid):
		return h.responseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Invalid submission data")
	case errors.Is(err, model.ErrSubmissionNotFound):
		return h.responseBuilder.BuildErrorResponse(c, http.StatusNotFound, "Submission not found")
	default:
		return h.responseBuilder.BuildErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to process form submission",
		)
	}
}

// HandleError handles validation errors
func (h *FormErrorHandlerImpl) HandleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrFormTitleRequired):
		return h.responseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form title is required")
	case errors.Is(err, model.ErrFormInvalid):
		return h.responseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form validation failed")
	default:
		return h.responseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}
}

// HandleOwnershipError handles ownership and authorization errors
func (h *FormErrorHandlerImpl) HandleOwnershipError(c echo.Context, err error) error {
	switch {
	case domainerrors.IsForbiddenError(err):
		return h.responseBuilder.BuildErrorResponse(
			c,
			http.StatusForbidden,
			"You don't have permission to access this resource",
		)
	case domainerrors.IsAuthenticationError(err):
		return h.responseBuilder.BuildErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	default:
		return h.responseBuilder.BuildErrorResponse(c, http.StatusForbidden, "Access denied")
	}
}

// HandleFormNotFoundError handles form not found errors
func (h *FormErrorHandlerImpl) HandleFormNotFoundError(c echo.Context, formID string) error {
	message := "Form not found"
	if formID != "" {
		message = fmt.Sprintf("Form not found: %s", formID)
	}

	return h.responseBuilder.BuildErrorResponse(c, http.StatusNotFound, message)
}

// HandleFormAccessError handles form access errors
func (h *FormErrorHandlerImpl) HandleFormAccessError(c echo.Context, err error) error {
	switch {
	case domainerrors.IsNotFound(err):
		return h.HandleFormNotFoundError(c, "")
	case domainerrors.IsForbiddenError(err):
		return h.HandleOwnershipError(c, err)
	case domainerrors.IsAuthenticationError(err):
		return h.HandleOwnershipError(c, err)
	default:
		return h.responseBuilder.BuildErrorResponse(c, http.StatusInternalServerError, "Failed to access form")
	}
}
