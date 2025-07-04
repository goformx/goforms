package openapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/openapi"
)

// OpenAPIValidationMiddleware validates requests and responses against OpenAPI specification
type OpenAPIValidationMiddleware struct {
	requestValidator  RequestValidator
	responseValidator ResponseValidator
	errorHandler      ValidationErrorHandler
	skipChecker       SkipConditionChecker
	routeCache        RouteCache
	responseCapture   ResponseCapture
	logger            logging.Logger
	config            *Config
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

	// Create components
	requestValidator := NewOpenAPIRequestValidator(router)
	responseValidator := NewOpenAPIResponseValidator(router)
	errorHandler := NewValidationErrorHandler(logger, config)
	skipChecker := NewSkipConditionChecker(config)
	routeCache := NewRouteCache()
	responseCapture := NewResponseCapture()

	return &OpenAPIValidationMiddleware{
		requestValidator:  requestValidator,
		responseValidator: responseValidator,
		errorHandler:      errorHandler,
		skipChecker:       skipChecker,
		routeCache:        routeCache,
		responseCapture:   responseCapture,
		logger:            logger,
		config:            config,
	}, nil
}

// Middleware returns the Echo middleware function
func (m *OpenAPIValidationMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.skipChecker.ShouldSkip(c.Path(), c.Request().Method) {
				return next(c)
			}

			route, pathParams, err := m.getOrFindRoute(c)
			if err != nil {
				return err
			}

			if err := m.validateRequestIfEnabled(c, route, pathParams); err != nil {
				return err
			}

			responseCapture := m.responseCapture.Setup(c)

			err = next(c)
			if err == nil {
				if validationErr := m.validateResponseIfEnabled(c, responseCapture, route, pathParams); validationErr != nil {
					return validationErr
				}
			}

			m.responseCapture.Restore(c, responseCapture)

			return err
		}
	}
}

// getOrFindRoute gets cached route or finds a new one
func (m *OpenAPIValidationMiddleware) getOrFindRoute(c echo.Context) (*routers.Route, map[string]string, error) {
	route, pathParams, ok := m.routeCache.Get(c)
	if !ok {
		var err error

		route, pathParams, err = m.findRoute(c.Request())
		if err != nil {
			return nil, nil, m.errorHandler.HandleError(c.Request().Context(), err, RequestValidationError, map[string]interface{}{
				"path":   c.Path(),
				"method": c.Request().Method,
				"ip":     c.RealIP(),
			})
		}

		m.routeCache.Set(c, route, pathParams)
	}

	return route, pathParams, nil
}

// validateRequestIfEnabled validates the request if enabled
func (m *OpenAPIValidationMiddleware) validateRequestIfEnabled(c echo.Context, route *routers.Route, pathParams map[string]string) error {
	if !m.config.EnableRequestValidation {
		return nil
	}

	if err := m.requestValidator.ValidateRequest(c.Request(), route, pathParams); err != nil {
		return m.errorHandler.HandleError(c.Request().Context(), err, RequestValidationError, map[string]interface{}{
			"path":   c.Path(),
			"method": c.Request().Method,
			"ip":     c.RealIP(),
		})
	}

	return nil
}

// validateResponseIfEnabled validates the response if enabled
func (m *OpenAPIValidationMiddleware) validateResponseIfEnabled(c echo.Context, responseCapture *CapturedResponse, route *routers.Route, pathParams map[string]string) error {
	if !m.config.EnableResponseValidation || responseCapture == nil {
		return nil
	}

	// Create a mock http.Response from Echo response for validation
	mockResponse := &http.Response{
		StatusCode: c.Response().Status,
		Header:     c.Response().Header(),
	}
	if validationErr := m.responseValidator.ValidateResponse(
		c.Request(),
		mockResponse,
		*responseCapture.Body,
		route,
		pathParams,
	); validationErr != nil {
		return m.errorHandler.HandleError(
			c.Request().Context(),
			validationErr,
			ResponseValidationError,
			map[string]interface{}{
				"path":   c.Path(),
				"method": c.Request().Method,
				"status": c.Response().Status,
			},
		)
	}

	return nil
}

// findRoute finds a route in the OpenAPI spec
func (m *OpenAPIValidationMiddleware) findRoute(req *http.Request) (*routers.Route, map[string]string, error) {
	// This would need to be implemented based on the router interface
	// For now, we'll need to access the router from the validators
	return nil, nil, fmt.Errorf("route finding not implemented in this refactor")
}

// Router returns the router for testing purposes
func (m *OpenAPIValidationMiddleware) Router() routers.Router {
	// This would need to be accessed from the validators
	return nil
}
