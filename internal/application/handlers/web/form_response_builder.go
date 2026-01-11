package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// FormResponseBuilderImpl implements FormResponseBuilder
type FormResponseBuilderImpl struct{}

// NewFormResponseBuilder creates a new form response builder
func NewFormResponseBuilder() FormResponseBuilder {
	return &FormResponseBuilderImpl{}
}

// BuildSuccessResponse builds a standardized success response
func (b *FormResponseBuilderImpl) BuildSuccessResponse(c echo.Context, message string, data map[string]any) error {
	return c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// BuildErrorResponse builds a standardized error response
func (b *FormResponseBuilderImpl) BuildErrorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, response.APIResponse{
		Success: false,
		Message: message,
	})
}

// BuildSchemaResponse builds a schema response
func (b *FormResponseBuilderImpl) BuildSchemaResponse(c echo.Context, schema model.JSON) error {
	// Set content type for JSON response
	c.Response().Header().Set("Content-Type", "application/json")

	// Wrap schema in standard API response format for frontend compatibility
	return c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Data:    schema,
	})
}

// BuildSubmissionResponse builds a submission response
func (b *FormResponseBuilderImpl) BuildSubmissionResponse(c echo.Context, submission *model.FormSubmission) error {
	return c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Form submitted successfully",
		Data: map[string]any{
			"submission_id": submission.ID,
			"status":        submission.Status,
			"submitted_at":  submission.SubmittedAt.Format(time.RFC3339),
		},
	})
}

// BuildFormResponse builds a form response
func (b *FormResponseBuilderImpl) BuildFormResponse(c echo.Context, form *model.Form) error {
	return c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Data: map[string]any{
			"form": map[string]any{
				"id":          form.ID,
				"title":       form.Title,
				"description": form.Description,
				"status":      form.Status,
				"schema":      form.Schema,
				"created_at":  form.CreatedAt.Format(time.RFC3339),
				"updated_at":  form.UpdatedAt.Format(time.RFC3339),
			},
		},
	})
}

// BuildFormListResponse builds a form list response
func (b *FormResponseBuilderImpl) BuildFormListResponse(c echo.Context, forms []*model.Form) error {
	formData := make([]map[string]any, len(forms))
	for i, form := range forms {
		formData[i] = map[string]any{
			"id":          form.ID,
			"title":       form.Title,
			"description": form.Description,
			"status":      form.Status,
			"created_at":  form.CreatedAt.Format(time.RFC3339),
			"updated_at":  form.UpdatedAt.Format(time.RFC3339),
		}
	}

	return c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Data: map[string]any{
			"forms": formData,
			"count": len(forms),
		},
	})
}

// BuildSubmissionListResponse builds a response for form submission lists
func (b *FormResponseBuilderImpl) BuildSubmissionListResponse(
	c echo.Context,
	submissions []*model.FormSubmission,
) error {
	submissionData := make([]map[string]any, len(submissions))
	for i, submission := range submissions {
		submissionData[i] = map[string]any{
			"id":           submission.ID,
			"form_id":      submission.FormID,
			"status":       submission.Status,
			"submitted_at": submission.SubmittedAt.Format(time.RFC3339),
			"data":         submission.Data,
		}
	}

	return c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Data: map[string]any{
			"submissions": submissionData,
			"count":       len(submissions),
		},
	})
}

// BuildValidationErrorResponse builds a validation error response
func (b *FormResponseBuilderImpl) BuildValidationErrorResponse(c echo.Context, field, message string) error {
	return c.JSON(http.StatusBadRequest, response.APIResponse{
		Success: false,
		Message: "Validation failed",
		Data: map[string]any{
			"field":   field,
			"message": message,
		},
	})
}

// BuildMultipleErrorResponse builds a response for multiple validation errors
func (b *FormResponseBuilderImpl) BuildMultipleErrorResponse(
	c echo.Context,
	errors []validation.Error,
) error {
	errorData := make([]map[string]any, len(errors))
	for i, err := range errors {
		errorData[i] = map[string]any{
			"field":   err.Field,
			"message": err.Message,
			"rule":    err.Rule,
		}
	}

	return c.JSON(http.StatusBadRequest, response.APIResponse{
		Success: false,
		Message: "Validation failed",
		Data: map[string]any{
			"errors": errorData,
		},
	})
}

// BuildNotFoundResponse builds a not found response
func (b *FormResponseBuilderImpl) BuildNotFoundResponse(c echo.Context, resource string) error {
	return c.JSON(http.StatusNotFound, response.APIResponse{
		Success: false,
		Message: resource + " not found",
	})
}

// BuildForbiddenResponse builds a forbidden response
func (b *FormResponseBuilderImpl) BuildForbiddenResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, response.APIResponse{
		Success: false,
		Message: message,
	})
}
