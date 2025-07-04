package http

import (
	"context"

	"github.com/goformx/goforms/internal/application/dto"
)

// RequestAdapter defines the interface for parsing HTTP requests to DTOs
type RequestAdapter interface {
	// Auth requests
	ParseLoginRequest(ctx Context) (*dto.LoginRequest, error)
	ParseSignupRequest(ctx Context) (*dto.SignupRequest, error)
	ParseLogoutRequest(ctx Context) (*dto.LogoutRequest, error)

	// Form requests
	ParseCreateFormRequest(ctx Context) (*dto.CreateFormRequest, error)
	ParseUpdateFormRequest(ctx Context) (*dto.UpdateFormRequest, error)
	ParseDeleteFormRequest(ctx Context) (*dto.DeleteFormRequest, error)
	ParseSubmitFormRequest(ctx Context) (*dto.SubmitFormRequest, error)
	ParsePaginationRequest(ctx Context) (*dto.PaginationRequest, error)

	// Utility methods
	ParseFormID(ctx Context) (string, error)
	ParseUserID(ctx Context) (string, error)
}

// ResponseAdapter defines the interface for converting DTOs to HTTP responses
type ResponseAdapter interface {
	// Auth responses
	BuildLoginResponse(ctx Context, response *dto.LoginResponse) error
	BuildSignupResponse(ctx Context, response *dto.SignupResponse) error
	BuildLogoutResponse(ctx Context, response *dto.LogoutResponse) error

	// Form responses
	BuildFormResponse(ctx Context, response *dto.FormResponse) error
	BuildFormListResponse(ctx Context, response *dto.FormListResponse) error
	BuildFormSchemaResponse(ctx Context, response *dto.FormSchemaResponse) error
	BuildSubmitFormResponse(ctx Context, response *dto.SubmitFormResponse) error

	// Error responses
	BuildErrorResponse(ctx Context, err error) error
	BuildValidationErrorResponse(ctx Context, errors []dto.ValidationError) error
	BuildNotFoundResponse(ctx Context, resource string) error
	BuildUnauthorizedResponse(ctx Context) error
	BuildForbiddenResponse(ctx Context) error

	// Generic responses
	BuildSuccessResponse(ctx Context, message string, data any) error
	BuildJSONResponse(ctx Context, statusCode int, data any) error
}

// Context represents a framework-agnostic HTTP context
type Context interface {
	// Request methods
	Method() string
	Path() string
	Param(name string) string
	QueryParam(name string) string
	FormValue(name string) string
	Body() []byte
	Headers() map[string]string

	// Response methods
	JSON(statusCode int, data any) error
	Redirect(statusCode int, url string) error
	NoContent(statusCode int) error

	// Context methods
	Get(key string) any
	Set(key string, value any)

	// Context propagation (needed for application services)
	RequestContext() context.Context
}
