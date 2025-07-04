package openapi

import (
	"context"
	"net/http"

	"github.com/getkin/kin-openapi/routers"
	"github.com/labstack/echo/v4"
)

// Config holds configuration for OpenAPI validation middleware
type Config struct {
	// EnableRequestValidation enables validation of incoming requests
	EnableRequestValidation bool
	// EnableResponseValidation enables validation of outgoing responses
	EnableResponseValidation bool
	// LogValidationErrors logs validation errors (doesn't block requests)
	LogValidationErrors bool
	// BlockInvalidRequests blocks requests that don't match the spec
	BlockInvalidRequests bool
	// BlockInvalidResponses blocks responses that don't match the spec
	BlockInvalidResponses bool
	// SkipPaths are paths that should be skipped for validation
	SkipPaths []string
	// SkipMethods are HTTP methods that should be skipped for validation
	SkipMethods []string
}

// ValidationErrorType represents the type of validation error
type ValidationErrorType string

const (
	RequestValidationError  ValidationErrorType = "request"
	ResponseValidationError ValidationErrorType = "response"
)

// context keys for caching route and pathParams
const (
	openapiRouteKey      = "openapi_route"
	openapiPathParamsKey = "openapi_path_params"
)

// RequestValidator validates incoming requests against OpenAPI spec
type RequestValidator interface {
	ValidateRequest(
		req *http.Request,
		route *routers.Route,
		pathParams map[string]string,
	) error
}

// ResponseValidator validates outgoing responses against OpenAPI spec
type ResponseValidator interface {
	ValidateResponse(
		req *http.Request,
		resp *http.Response,
		body []byte,
		route *routers.Route,
		pathParams map[string]string,
	) error
}

// ValidationErrorHandler handles validation errors consistently
type ValidationErrorHandler interface {
	HandleError(
		ctx context.Context,
		err error,
		errorType ValidationErrorType,
		metadata map[string]interface{},
	) error
}

// SkipConditionChecker determines if validation should be skipped
type SkipConditionChecker interface {
	ShouldSkip(path, method string) bool
}

// RouteCache manages route and parameter caching
type RouteCache interface {
	Get(c echo.Context) (*routers.Route, map[string]string, bool)
	Set(c echo.Context, route *routers.Route, pathParams map[string]string)
}

// ResponseCapture manages response body capture for validation
type ResponseCapture interface {
	Setup(c echo.Context) *CapturedResponse
	Restore(c echo.Context, capture *CapturedResponse)
}

// CapturedResponse holds information about captured response
type CapturedResponse struct {
	Body           *[]byte
	OriginalWriter http.ResponseWriter
}
