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

			if validationErr := m.validateRequestIfEnabled(c, route, pathParams); validationErr != nil {
				return validationErr
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
	if ok {
		return route, pathParams, nil
	}

	return m.findAndCacheRoute(c)
}

// findAndCacheRoute finds a route and caches it
func (m *OpenAPIValidationMiddleware) findAndCacheRoute(c echo.Context) (*routers.Route, map[string]string, error) {
	requestValidator, ok := m.requestValidator.(*OpenAPIRequestValidator)
	if !ok {
		return nil, nil, fmt.Errorf("request validator does not support route finding")
	}

	route, pathParams, err := requestValidator.FindRoute(c.Request())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find route: %w", m.errorHandler.HandleError(
			c.Request().Context(),
			err,
			RequestValidationError,
			map[string]interface{}{
				"path":   c.Path(),
				"method": c.Request().Method,
				"ip":     c.RealIP(),
			},
		))
	}

	m.routeCache.Set(c, route, pathParams)

	return route, pathParams, nil
}

// validateRequestIfEnabled validates the request if enabled
func (m *OpenAPIValidationMiddleware) validateRequestIfEnabled(
	c echo.Context,
	route *routers.Route,
	pathParams map[string]string,
) error {
	if !m.config.EnableRequestValidation {
		return nil
	}

	if err := m.requestValidator.ValidateRequest(c.Request(), route, pathParams); err != nil {
		return fmt.Errorf("request validation failed: %w", m.errorHandler.HandleError(
			c.Request().Context(),
			err,
			RequestValidationError,
			map[string]interface{}{
				"path":   c.Path(),
				"method": c.Request().Method,
				"ip":     c.RealIP(),
			},
		))
	}

	return nil
}

// validateResponseIfEnabled validates the response if enabled
func (m *OpenAPIValidationMiddleware) validateResponseIfEnabled(
	c echo.Context,
	responseCapture *CapturedResponse,
	route *routers.Route,
	pathParams map[string]string,
) error {
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
		return fmt.Errorf("response validation failed: %w", m.errorHandler.HandleError(
			c.Request().Context(),
			validationErr,
			ResponseValidationError,
			map[string]interface{}{
				"path":   c.Path(),
				"method": c.Request().Method,
				"status": c.Response().Status,
			},
		))
	}

	return nil
}

// Router returns the router for testing purposes
func (m *OpenAPIValidationMiddleware) Router() routers.Router {
	// Access router from the request validator
	if requestValidator, ok := m.requestValidator.(*OpenAPIRequestValidator); ok {
		return requestValidator.Router
	}

	return nil
}
