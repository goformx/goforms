package openapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

// OpenAPIRequestValidator implements RequestValidator
type OpenAPIRequestValidator struct {
	Router routers.Router
}

// NewOpenAPIRequestValidator creates a new request validator
func NewOpenAPIRequestValidator(router routers.Router) RequestValidator {
	return &OpenAPIRequestValidator{Router: router}
}

// ValidateRequest validates the incoming request against the OpenAPI spec
func (v *OpenAPIRequestValidator) ValidateRequest(
	req *http.Request,
	route *routers.Route,
	pathParams map[string]string,
) error {
	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			IncludeResponseStatus: true,
			AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
				if input.SecurityScheme == nil {
					return fmt.Errorf("security scheme is nil for %s", input.SecuritySchemeName)
				}
				if input.SecuritySchemeName == "SessionAuth" {
					// For test purposes, always succeed
					return nil
				}

				return fmt.Errorf("unsupported security scheme: %s", input.SecuritySchemeName)
			},
		},
	}
	if validateErr := openapi3filter.ValidateRequest(context.Background(), validationInput); validateErr != nil {
		return fmt.Errorf("request validation failed: %w", validateErr)
	}

	return nil
}

// OpenAPIResponseValidator implements ResponseValidator
type OpenAPIResponseValidator struct {
	router routers.Router
}

// NewOpenAPIResponseValidator creates a new response validator
func NewOpenAPIResponseValidator(router routers.Router) ResponseValidator {
	return &OpenAPIResponseValidator{router: router}
}

// ValidateResponse validates the outgoing response against the OpenAPI spec
func (v *OpenAPIResponseValidator) ValidateResponse(
	req *http.Request,
	resp *http.Response,
	body []byte,
	route *routers.Route,
	pathParams map[string]string,
) error {
	validationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				IncludeResponseStatus: true,
			},
		},
		Status: resp.StatusCode,
		Header: resp.Header,
		Body:   io.NopCloser(bytes.NewReader(body)),
	}

	if validateErr := openapi3filter.ValidateResponse(context.Background(), validationInput); validateErr != nil {
		return fmt.Errorf("response validation failed: %w", validateErr)
	}

	return nil
}

// FindRoute finds a route in the OpenAPI spec
func (v *OpenAPIRequestValidator) FindRoute(req *http.Request) (*routers.Route, map[string]string, error) {
	route, pathParams, err := v.Router.FindRoute(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find route: %w", err)
	}

	return route, pathParams, nil
}
