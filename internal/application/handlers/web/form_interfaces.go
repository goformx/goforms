package web

import (
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/labstack/echo/v4"
)

// FormCreateRequest represents the data needed to create a form
type FormCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CorsOrigins string `json:"cors_origins"`
}

// FormUpdateRequest represents the data needed to update a form
type FormUpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CorsOrigins string `json:"cors_origins"`
}

// FormRetriever interface for retrieving forms
type FormRetriever interface {
	GetFormByID(c echo.Context) (*model.Form, error)
	GetFormWithOwnership(c echo.Context) (*model.Form, error)
}

// FormOwnershipValidator interface for validating form ownership
type FormOwnershipValidator interface {
	RequireFormOwnership(c echo.Context, form *model.Form) error
}

// FormRequestProcessor interface for processing form requests
type FormRequestProcessor interface {
	ProcessCreateRequest(c echo.Context) (*FormCreateRequest, error)
	ProcessUpdateRequest(c echo.Context) (*FormUpdateRequest, error)
	ProcessSchemaUpdateRequest(c echo.Context) (model.JSON, error)
	ProcessSubmissionRequest(c echo.Context) (model.JSON, error)
}

// FormResponseBuilder interface for building standardized responses
type FormResponseBuilder interface {
	BuildSuccessResponse(c echo.Context, message string, data map[string]any) error
	BuildErrorResponse(c echo.Context, statusCode int, message string) error
	BuildSchemaResponse(c echo.Context, schema model.JSON) error
	BuildSubmissionResponse(c echo.Context, submission *model.FormSubmission) error
}

// FormErrorHandler interface for handling form-specific errors
type FormErrorHandler interface {
	HandleSchemaError(c echo.Context, err error) error
	HandleSubmissionError(c echo.Context, err error) error
	HandleValidationError(c echo.Context, err error) error
	HandleOwnershipError(c echo.Context, err error) error
}
