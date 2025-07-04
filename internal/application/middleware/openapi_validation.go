package middleware

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/openapi"
)

// context keys for caching route and pathParams
const (
	openapiRouteKey      = "openapi_route"
	openapiPathParamsKey = "openapi_path_params"
)

// OpenAPIValidationMiddleware validates requests and responses against OpenAPI specification
type OpenAPIValidationMiddleware struct {
	doc    *openapi3.T
	router routers.Router
	logger logging.Logger
	config *Config
}

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

// NewOpenAPIValidationMiddleware creates a new OpenAPI validation middleware
func NewOpenAPIValidationMiddleware(logger logging.Logger, config *Config) (*OpenAPIValidationMiddleware, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Load and parse the OpenAPI specification
	loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}

	doc, err := loader.LoadFromData([]byte(openapi.OpenAPISpec))
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the specification
	if validateErr := doc.Validate(loader.Context); validateErr != nil {
		return nil, fmt.Errorf("invalid OpenAPI specification: %w", validateErr)
	}

	// Create router for path/method lookup using gorillamux router
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	return &OpenAPIValidationMiddleware{
		doc:    doc,
		router: router,
		logger: logger,
		config: config,
	}, nil
}

// Middleware returns the Echo middleware function
func (m *OpenAPIValidationMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.shouldSkip(c) {
				return next(c)
			}

			// Find and cache route/params if not already cached
			route, pathParams, ok := getOpenAPIRouteAndParams(c)
			if !ok {
				var err error

				route, pathParams, err = m.router.FindRoute(c.Request())
				if err != nil {
					return m.handleRequestValidationError(c, err)
				}

				setOpenAPIRouteAndParams(c, route, pathParams)
			}

			// Validate request if enabled
			if m.config.EnableRequestValidation {
				if err := m.validateRequestWithRoute(c, route, pathParams); err != nil {
					return m.handleRequestValidationError(c, err)
				}
			}

			// Capture response for validation
			responseCapture := m.setupResponseCapture(c)
			err := next(c)

			// Validate response if enabled
			if err == nil && m.config.EnableResponseValidation && responseCapture != nil {
				if validationErr := m.validateResponseWithRoute(c, *responseCapture.body, route, pathParams); validationErr != nil {
					return m.handleResponseValidationError(c, validationErr, responseCapture)
				}
			}

			m.restoreResponseWriter(c, responseCapture)

			return err
		}
	}
}

// validateRequestIfEnabled validates the request if request validation is enabled
func (m *OpenAPIValidationMiddleware) validateRequestIfEnabled(c echo.Context) error {
	if !m.config.EnableRequestValidation {
		return nil
	}

	if err := m.validateRequest(c); err != nil {
		return m.handleRequestValidationError(c, err)
	}

	return nil
}

// handleRequestValidationError handles request validation errors based on configuration
func (m *OpenAPIValidationMiddleware) handleRequestValidationError(c echo.Context, err error) error {
	if m.config.BlockInvalidRequests {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Request validation failed: %v", err))
	}

	if m.config.LogValidationErrors {
		m.logger.Warn("Request validation failed",
			"error", err,
			"path", c.Path(),
			"method", c.Request().Method,
			"ip", c.RealIP(),
		)
	}

	return nil
}

// setupResponseCapture sets up response capture for validation
func (m *OpenAPIValidationMiddleware) setupResponseCapture(c echo.Context) *responseCapture {
	if !m.config.EnableResponseValidation {
		return nil
	}

	responseBody := make([]byte, 0)
	originalWriter := c.Response().Writer

	responseWriter := &responseCaptureWriter{
		Response: c.Response(),
		Writer:   originalWriter,
		body:     &responseBody,
	}
	c.Response().Writer = responseWriter

	return &responseCapture{
		body:           &responseBody,
		originalWriter: originalWriter,
	}
}

// validateResponseIfEnabled validates the response if response validation is enabled
func (m *OpenAPIValidationMiddleware) validateResponseIfEnabled(c echo.Context, capture *responseCapture) error {
	if !m.config.EnableResponseValidation || capture == nil {
		return nil
	}

	if validationErr := m.validateResponse(c, *capture.body); validationErr != nil {
		return m.handleResponseValidationError(c, validationErr, capture)
	}

	return nil
}

// handleResponseValidationError handles response validation errors based on configuration
func (m *OpenAPIValidationMiddleware) handleResponseValidationError(
	c echo.Context,
	err error,
	capture *responseCapture,
) error {
	if m.config.BlockInvalidResponses {
		// Restore original writer and return error
		if capture != nil && capture.originalWriter != nil {
			c.Response().Writer = capture.originalWriter
		}

		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Sprintf("Response validation failed: %v", err))
	}

	if m.config.LogValidationErrors {
		m.logger.Warn("Response validation failed",
			"error", err,
			"path", c.Path(),
			"method", c.Request().Method,
			"status", c.Response().Status,
		)
	}

	return nil
}

// restoreResponseWriter restores the original response writer
func (m *OpenAPIValidationMiddleware) restoreResponseWriter(c echo.Context, capture *responseCapture) {
	if capture != nil && capture.originalWriter != nil {
		c.Response().Writer = capture.originalWriter
	}
}

// responseCapture holds information about captured response
type responseCapture struct {
	body           *[]byte
	originalWriter http.ResponseWriter
}

// shouldSkip checks if validation should be skipped for this request
func (m *OpenAPIValidationMiddleware) shouldSkip(c echo.Context) bool {
	path := c.Path()
	method := c.Request().Method

	// Check skip paths
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// Check skip methods
	for _, skipMethod := range m.config.SkipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}

// validateRequest validates the incoming request against the OpenAPI spec
func (m *OpenAPIValidationMiddleware) validateRequest(c echo.Context) error {
	request := c.Request()

	// Find the route in the OpenAPI spec
	route, pathParams, err := m.router.FindRoute(request)
	if err != nil {
		return fmt.Errorf("route not found in OpenAPI spec: %w", err)
	}

	// Create validation input
	validationInput := &openapi3filter.RequestValidationInput{
		Request:    request,
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

	// Validate the request
	if validateErr := openapi3filter.ValidateRequest(context.Background(), validationInput); validateErr != nil {
		return fmt.Errorf("request validation failed: %w", validateErr)
	}

	return nil
}

// validateResponse validates the outgoing response against the OpenAPI spec
func (m *OpenAPIValidationMiddleware) validateResponse(c echo.Context, responseBody []byte) error {
	request := c.Request()
	response := c.Response()

	// Find the route in the OpenAPI spec
	route, pathParams, err := m.router.FindRoute(request)
	if err != nil {
		return fmt.Errorf("route not found in OpenAPI spec: %w", err)
	}

	// Create validation input
	validationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    request,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				IncludeResponseStatus: true,
			},
		},
		Status: response.Status,
		Header: response.Header(),
		Body:   io.NopCloser(bytes.NewReader(responseBody)),
	}

	// Validate the response
	if validateErr := openapi3filter.ValidateResponse(context.Background(), validationInput); validateErr != nil {
		return fmt.Errorf("response validation failed: %w", validateErr)
	}

	return nil
}

// responseCaptureWriter captures the response body for validation
type responseCaptureWriter struct {
	Response *echo.Response
	Writer   http.ResponseWriter
	body     *[]byte
}

// Write captures the response body
func (w *responseCaptureWriter) Write(b []byte) (int, error) {
	*w.body = append(*w.body, b...)

	return w.Writer.Write(b)
}

// WriteHeader captures the response status
func (w *responseCaptureWriter) WriteHeader(statusCode int) {
	w.Response.Status = statusCode
	w.Writer.WriteHeader(statusCode)
}

// Header returns the response headers
func (w *responseCaptureWriter) Header() http.Header {
	return w.Writer.Header()
}

// Flush flushes the underlying writer
func (w *responseCaptureWriter) Flush() {
	if flusher, ok := w.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack hijacks the connection
func (w *responseCaptureWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.Writer.(http.Hijacker); ok {
		return hijacker.Hijack()
	}

	return nil, nil, fmt.Errorf("underlying writer does not implement http.Hijacker")
}

// Router returns the router for testing purposes
func (m *OpenAPIValidationMiddleware) Router() routers.Router {
	return m.router
}

// Helper to set route and pathParams in context
func setOpenAPIRouteAndParams(c echo.Context, route *routers.Route, pathParams map[string]string) {
	c.Set(openapiRouteKey, route)
	c.Set(openapiPathParamsKey, pathParams)
}

// Helper to get route and pathParams from context
func getOpenAPIRouteAndParams(c echo.Context) (*routers.Route, map[string]string, bool) {
	r, rok := c.Get(openapiRouteKey).(*routers.Route)
	p, pok := c.Get(openapiPathParamsKey).(map[string]string)

	return r, p, rok && pok
}

// New versions that use cached route/params
func (m *OpenAPIValidationMiddleware) validateRequestWithRoute(c echo.Context, route *routers.Route, pathParams map[string]string) error {
	request := c.Request()

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    request,
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

func (m *OpenAPIValidationMiddleware) validateResponseWithRoute(c echo.Context, responseBody []byte, route *routers.Route, pathParams map[string]string) error {
	request := c.Request()
	response := c.Response()

	validationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    request,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				IncludeResponseStatus: true,
			},
		},
		Status: response.Status,
		Header: response.Header(),
		Body:   io.NopCloser(bytes.NewReader(responseBody)),
	}
	if validateErr := openapi3filter.ValidateResponse(context.Background(), validationInput); validateErr != nil {
		return fmt.Errorf("response validation failed: %w", validateErr)
	}

	return nil
}
