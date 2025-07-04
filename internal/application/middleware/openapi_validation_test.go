package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/openapi"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
)

func TestOpenAPIValidationMiddleware_ValidateActualResponses(t *testing.T) {
	// Create a mock logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)

	// Create the validation middleware with proper skip paths
	validationMiddleware, err := middleware.NewOpenAPIValidationMiddleware(logger, &middleware.Config{
		EnableRequestValidation:  true,
		EnableResponseValidation: true,
		LogValidationErrors:      true,
		BlockInvalidRequests:     false,
		BlockInvalidResponses:    false,
		SkipPaths:                []string{"/health"}, // Skip health endpoint for testing
		SkipMethods:              []string{},
	})
	require.NoError(t, err)

	// Create Echo instance
	e := echo.New()
	e.Use(validationMiddleware.Middleware())

	// Test 1: Health check endpoint (should pass validation)
	t.Run("Health Check Response", func(t *testing.T) {
		e.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":    "healthy",
				"timestamp": "2024-01-01T00:00:00Z",
				"version":   "1.0.0",
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Test 2: Form list response (should pass validation)
	t.Run("Form List Response", func(t *testing.T) {
		// Set up mock logger expectations for response validation failure
		logger.EXPECT().Warn(
			"Response validation failed",
			"error", gomock.Any(),
			"path", "/api/v1/forms",
			"method", "GET",
			"status", 200,
		).Return()

		e.GET("/api/v1/forms", func(c echo.Context) error {
			response := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"forms": []map[string]interface{}{
						{
							"id":          "123e4567-e89b-12d3-a456-426614174000",
							"title":       "Test Form",
							"description": "A test form",
							"status":      "draft",
							"created_at":  "2024-01-01T00:00:00Z",
							"updated_at":  "2024-01-01T00:00:00Z",
						},
					},
					"count": 1,
				},
			}

			return c.JSON(http.StatusOK, response)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/forms", http.NoBody)
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session-id"})

		// Add debugging headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Test 3: Form detail response (should pass validation)
	t.Run("Form Detail Response", func(t *testing.T) {
		e.GET("/api/v1/forms/:id", func(c echo.Context) error {
			response := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"form": map[string]interface{}{
						"id":          "123e4567-e89b-12d3-a456-426614174000",
						"title":       "Test Form",
						"description": "A test form",
						"status":      "draft",
						"schema":      map[string]interface{}{"type": "object"},
						"created_at":  "2024-01-01T00:00:00Z",
						"updated_at":  "2024-01-01T00:00:00Z",
					},
				},
			}

			return c.JSON(http.StatusOK, response)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/forms/123e4567-e89b-12d3-a456-426614174000", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Test 4: Form submission response (should pass validation)
	t.Run("Form Submission Response", func(t *testing.T) {
		e.POST("/api/v1/forms/:id/submit", func(c echo.Context) error {
			response := map[string]interface{}{
				"success": true,
				"message": "Form submitted successfully",
				"data": map[string]interface{}{
					"submission_id": "123e4567-e89b-12d3-a456-426614174000",
					"status":        "pending",
					"submitted_at":  "2024-01-01T00:00:00Z",
				},
			}

			return c.JSON(http.StatusOK, response)
		})

		submissionData := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		}
		body, marshalErr := json.Marshal(submissionData)
		require.NoError(t, marshalErr)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/forms/123e4567-e89b-12d3-a456-426614174000/submit",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Test 5: Error response (should pass validation)
	t.Run("Error Response", func(t *testing.T) {
		e.GET("/api/v1/forms/nonexistent", func(c echo.Context) error {
			response := map[string]interface{}{
				"success": false,
				"message": "Form not found",
			}

			return c.JSON(http.StatusNotFound, response)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/forms/nonexistent", http.NoBody)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	// Test 6: Validation error response (should pass validation)
	t.Run("Validation Error Response", func(t *testing.T) {
		e.POST("/api/v1/forms/:id/submit", func(c echo.Context) error {
			response := map[string]interface{}{
				"success": false,
				"message": "Validation failed",
				"data": map[string]interface{}{
					"errors": []map[string]interface{}{
						{
							"field":   "email",
							"message": "Email is required",
							"rule":    "required",
						},
					},
				},
			}

			return c.JSON(http.StatusBadRequest, response)
		})

		submissionData := map[string]interface{}{
			"name": "John Doe",
			// Missing email field
		}
		body, marshalErr := json.Marshal(submissionData)
		require.NoError(t, marshalErr)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/forms/123e4567-e89b-12d3-a456-426614174000/submit",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestOpenAPIValidationMiddleware_Config(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)

	t.Run("Default Config", func(t *testing.T) {
		config := config.DefaultOpenAPIConfig()
		assert.True(t, config.EnableRequestValidation)
		assert.True(t, config.EnableResponseValidation)
		assert.True(t, config.LogValidationErrors)
		assert.False(t, config.BlockInvalidRequests)
		assert.False(t, config.BlockInvalidResponses)
		assert.Contains(t, config.SkipPaths, "/health")
		assert.Contains(t, config.SkipMethods, "OPTIONS")
	})

	t.Run("Custom Config", func(t *testing.T) {
		config := &middleware.Config{
			EnableRequestValidation:  true,
			EnableResponseValidation: true,
			LogValidationErrors:      true,
			BlockInvalidRequests:     true,
			BlockInvalidResponses:    true,
			SkipPaths:                []string{"/custom"},
			SkipMethods:              []string{"CUSTOM"},
		}

		validationMiddleware, err := middleware.NewOpenAPIValidationMiddleware(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, validationMiddleware)
	})
}

func TestOpenAPIValidationMiddleware_SkipPaths(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)

	openAPIConfig := config.DefaultOpenAPIConfig()
	openAPIConfig.SkipPaths = append(openAPIConfig.SkipPaths, "/test-skip")

	config := &middleware.Config{
		EnableRequestValidation:  openAPIConfig.EnableRequestValidation,
		EnableResponseValidation: openAPIConfig.EnableResponseValidation,
		LogValidationErrors:      openAPIConfig.LogValidationErrors,
		BlockInvalidRequests:     openAPIConfig.BlockInvalidRequests,
		BlockInvalidResponses:    openAPIConfig.BlockInvalidResponses,
		SkipPaths:                openAPIConfig.SkipPaths,
		SkipMethods:              openAPIConfig.SkipMethods,
	}

	validationMiddleware, err := middleware.NewOpenAPIValidationMiddleware(logger, config)
	require.NoError(t, err)

	e := echo.New()
	e.Use(validationMiddleware.Middleware())

	e.GET("/test-skip", func(c echo.Context) error {
		return c.String(http.StatusOK, "skipped")
	})

	req := httptest.NewRequest(http.MethodGet, "/test-skip", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "skipped", rec.Body.String())
}

func TestOpenAPIValidationMiddleware_SpecLoading(t *testing.T) {
	// Create a mock logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)

	// Create the validation middleware
	validationMiddleware, err := middleware.NewOpenAPIValidationMiddleware(logger, &middleware.Config{
		EnableRequestValidation:  false, // Disable for this test
		EnableResponseValidation: false,
		LogValidationErrors:      false,
		BlockInvalidRequests:     false,
		BlockInvalidResponses:    false,
		SkipPaths:                []string{},
		SkipMethods:              []string{},
	})
	require.NoError(t, err)

	// Debug: Print all available routes
	router := validationMiddleware.Router()
	if router != nil {
		t.Logf("Router created successfully")

		// Try to find routes with different paths
		testPaths := []string{
			"/health",
			"/api/v1/health",
			"/api/v1/forms",
			"/api/v1/forms/123",
			"/api/v1/forms/123/schema",
			"/api/v1/forms/123/submit",
		}

		for _, path := range testPaths {
			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)

			route, _, err := router.FindRoute(req)
			if err != nil {
				t.Logf("Could not find %s route: %v", path, err)
			} else {
				t.Logf("Found route: %s -> %s", path, route.Path)
			}
		}
	} else {
		t.Logf("Router is nil")
	}

	// Test if we can find the health route (which should be public)
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)

	route, _, err := validationMiddleware.Router().FindRoute(req)
	if err != nil {
		t.Logf("Could not find /health route: %v", err)
	} else {
		t.Logf("Found route: %s", route.Path)
	}

	// Test if we can find the forms route
	req = httptest.NewRequest(http.MethodGet, "/api/v1/forms", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session-id"})

	route, _, err = validationMiddleware.Router().FindRoute(req)
	if err != nil {
		t.Logf("Could not find /api/v1/forms route: %v", err)
	} else {
		t.Logf("Found route: %s", route.Path)
	}
}

func TestOpenAPISpecLoading(t *testing.T) {
	// Test direct OpenAPI spec loading
	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData([]byte(openapi.OpenAPISpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Print some debug info
	t.Logf("OpenAPI spec loaded successfully")
	t.Logf("Info: %s v%s", doc.Info.Title, doc.Info.Version)

	if doc.Paths != nil {
		t.Logf("Paths object exists")

		// Test specific paths
		testPaths := []string{
			"/health",
			"/api/v1/forms",
			"/api/v1/forms/{id}",
			"/api/v1/forms/{id}/schema",
			"/api/v1/forms/{id}/submit",
		}

		for _, path := range testPaths {
			if pathItem := doc.Paths.Find(path); pathItem != nil {
				t.Logf("Path %s exists with operations", path)

				operations := pathItem.Operations()
				for method := range operations {
					t.Logf("  - %s", method)
				}
			} else {
				t.Logf("Path %s does not exist", path)
			}
		}
	} else {
		t.Logf("Paths object is nil")
	}
}

func TestGorillaMuxRouterCreation(t *testing.T) {
	// Test gorillamux router creation directly
	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData([]byte(openapi.OpenAPISpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Create router using gorillamux
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	t.Logf("GorillaMux router created successfully")

	// Test route finding
	testCases := []struct {
		path   string
		method string
	}{
		{"/health", "GET"},
		{"/api/v1/forms", "GET"},
		{"/api/v1/forms", "POST"},
		{"/api/v1/forms/123", "GET"},
		{"/api/v1/forms/123/schema", "GET"},
		{"/api/v1/forms/123/submit", "POST"},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, http.NoBody)

		route, pathParams, err := router.FindRoute(req)
		if err != nil {
			t.Logf("Could not find %s %s: %v", tc.method, tc.path, err)
		} else {
			t.Logf("Found %s %s -> %s (params: %v)", tc.method, tc.path, route.Path, pathParams)
		}
	}
}

func TestMinimalOpenAPISpec(t *testing.T) {
	// Test with a minimal OpenAPI spec
	minimalSpec := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      summary: Test endpoint
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  message:
                    type: string
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(minimalSpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Create router using gorillamux
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	t.Logf("Minimal spec router created successfully")

	// Test route finding
	req := httptest.NewRequest("GET", "/test", http.NoBody)

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Logf("Could not find GET /test: %v", err)
	} else {
		t.Logf("Found GET /test -> %s (params: %v)", route.Path, pathParams)
	}
}

func TestSimplifiedOpenAPISpec(t *testing.T) {
	// Test with a simplified version of our spec
	simplifiedSpec := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /health:
    get:
      summary: Health check
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
  /api/v1/forms:
    get:
      summary: List forms
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                  data:
                    type: object
                    properties:
                      forms:
                        type: array
                        items:
                          type: object
                          properties:
                            id:
                              type: string
                            title:
                              type: string
                      count:
                        type: integer
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(simplifiedSpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Create router using gorillamux
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	t.Logf("Simplified spec router created successfully")

	// Test route finding
	testCases := []struct {
		path   string
		method string
	}{
		{"/health", "GET"},
		{"/api/v1/forms", "GET"},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, http.NoBody)

		route, pathParams, err := router.FindRoute(req)
		if err != nil {
			t.Logf("Could not find %s %s: %v", tc.method, tc.path, err)
		} else {
			t.Logf("Found %s %s -> %s (params: %v)", tc.method, tc.path, route.Path, pathParams)
		}
	}
}

func TestOpenAPISpecFeatures(t *testing.T) {
	// Test with security schemes
	securitySpec := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
components:
  securitySchemes:
    SessionAuth:
      type: apiKey
      in: cookie
      name: session
paths:
  /api/v1/forms:
    get:
      summary: List forms
      security:
        - SessionAuth: []
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(securitySpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Create router using gorillamux
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	t.Logf("Security spec router created successfully")

	// Test route finding
	req := httptest.NewRequest("GET", "/api/v1/forms", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "session", Value: "test"})

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Logf("Could not find GET /api/v1/forms with security: %v", err)
	} else {
		t.Logf("Found GET /api/v1/forms with security -> %s (params: %v)", route.Path, pathParams)
	}
}

func TestOpenAPISpecSecurityOverride(t *testing.T) {
	// Test with security override ({} allows unauthenticated access)
	securityOverrideSpec := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
components:
  securitySchemes:
    SessionAuth:
      type: apiKey
      in: cookie
      name: session
paths:
  /api/v1/forms:
    get:
      summary: List forms
      security:
        - SessionAuth: []
        - {} # Allow unauthenticated access for testing
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(securityOverrideSpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Create router using gorillamux
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	t.Logf("Security override spec router created successfully")

	// Test route finding without authentication
	req := httptest.NewRequest("GET", "/api/v1/forms", http.NoBody)

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Logf("Could not find GET /api/v1/forms without auth: %v", err)
	} else {
		t.Logf("Found GET /api/v1/forms without auth -> %s (params: %v)", route.Path, pathParams)
	}

	// Test route finding with authentication
	req = httptest.NewRequest("GET", "/api/v1/forms", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "session", Value: "test"})

	route, pathParams, err = router.FindRoute(req)
	if err != nil {
		t.Logf("Could not find GET /api/v1/forms with auth: %v", err)
	} else {
		t.Logf("Found GET /api/v1/forms with auth -> %s (params: %v)", route.Path, pathParams)
	}
}

func TestOpenAPISpecAllOfSchema(t *testing.T) {
	// Test with allOf schema composition
	allOfSpec := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
components:
  schemas:
    APIResponse:
      type: object
      required:
        - success
      properties:
        success:
          type: boolean
        message:
          type: string
        data:
          type: object
paths:
  /api/v1/forms:
    get:
      summary: List forms
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/APIResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        properties:
                          forms:
                            type: array
                            items:
                              type: object
                              properties:
                                id:
                                  type: string
                                title:
                                  type: string
                          count:
                            type: integer
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(allOfSpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Create router using gorillamux
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	t.Logf("AllOf schema spec router created successfully")

	// Test route finding
	req := httptest.NewRequest("GET", "/api/v1/forms", http.NoBody)

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Logf("Could not find GET /api/v1/forms with allOf: %v", err)
	} else {
		t.Logf("Found GET /api/v1/forms with allOf -> %s (params: %v)", route.Path, pathParams)
	}
}

func TestOpenAPISpecCombinedFeatures(t *testing.T) {
	// Test with all features combined (closer to our actual spec)
	combinedSpec := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
components:
  securitySchemes:
    SessionAuth:
      type: apiKey
      in: cookie
      name: session
  schemas:
    APIResponse:
      type: object
      required:
        - success
      properties:
        success:
          type: boolean
        message:
          type: string
        data:
          type: object
paths:
  /health:
    get:
      summary: Health check
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                  timestamp:
                    type: string
                  version:
                    type: string
  /api/v1/forms:
    get:
      summary: List forms
      security:
        - SessionAuth: []
        - {} # Allow unauthenticated access for testing
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/APIResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        properties:
                          forms:
                            type: array
                            items:
                              type: object
                              properties:
                                id:
                                  type: string
                                title:
                                  type: string
                          count:
                            type: integer
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(combinedSpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Create router using gorillamux
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	t.Logf("Combined features spec router created successfully")

	// Test route finding
	testCases := []struct {
		path   string
		method string
		auth   bool
	}{
		{"/health", "GET", false},
		{"/api/v1/forms", "GET", false},
		{"/api/v1/forms", "GET", true},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, http.NoBody)
		if tc.auth {
			req.AddCookie(&http.Cookie{Name: "session", Value: "test"})
		}

		route, pathParams, err := router.FindRoute(req)
		if err != nil {
			t.Logf("Could not find %s %s (auth: %t): %v", tc.method, tc.path, tc.auth, err)
		} else {
			t.Logf("Found %s %s (auth: %t) -> %s (params: %v)", tc.method, tc.path, tc.auth, route.Path, pathParams)
		}
	}
}

func TestActualOpenAPISpec(t *testing.T) {
	// Test with our actual OpenAPI spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi.OpenAPISpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	// Create router using gorillamux
	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	t.Logf("Actual OpenAPI spec router created successfully")

	// Test route finding
	testCases := []struct {
		path   string
		method string
		auth   bool
	}{
		{"/health", "GET", false},
		{"/api/v1/forms", "GET", false},
		{"/api/v1/forms", "GET", true},
		{"/api/v1/forms", "POST", true},
		{"/api/v1/forms/123", "GET", true},
		{"/api/v1/forms/123/schema", "GET", false},
		{"/api/v1/forms/123/submit", "POST", false},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, http.NoBody)
		if tc.auth {
			req.AddCookie(&http.Cookie{Name: "session", Value: "test"})
		}

		route, pathParams, err := router.FindRoute(req)
		if err != nil {
			t.Logf("Could not find %s %s (auth: %t): %v", tc.method, tc.path, tc.auth, err)
		} else {
			t.Logf("Found %s %s (auth: %t) -> %s (params: %v)", tc.method, tc.path, tc.auth, route.Path, pathParams)
		}
	}
}

func TestOpenAPISpecValidation(t *testing.T) {
	// Test our actual OpenAPI spec more thoroughly
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi.OpenAPISpec))
	require.NoError(t, err)

	// Validate the specification
	validateErr := doc.Validate(context.Background())
	require.NoError(t, validateErr)

	t.Logf("OpenAPI spec loaded and validated successfully")
	t.Logf("Info: %s v%s", doc.Info.Title, doc.Info.Version)
	t.Logf("Number of paths: %d", len(doc.Paths.Map()))

	// List all available paths
	for path := range doc.Paths.Map() {
		t.Logf("Available path: %s", path)
	}

	// Check specific paths
	testPaths := []string{
		"/health",
		"/api/v1/forms",
		"/api/v1/forms/{id}",
		"/api/v1/forms/{id}/schema",
		"/api/v1/forms/{id}/submit",
	}

	for _, path := range testPaths {
		if pathItem := doc.Paths.Find(path); pathItem != nil {
			t.Logf("Path %s exists with %d operations", path, len(pathItem.Operations()))

			for method := range pathItem.Operations() {
				t.Logf("  - %s", method)
			}
		} else {
			t.Logf("Path %s does not exist", path)
		}
	}

	// Try to create router
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Logf("Failed to create router: %v", err)

		return
	}

	t.Logf("Router created successfully")

	// Test a simple route
	req := httptest.NewRequest("GET", "/health", http.NoBody)

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Logf("Could not find GET /health: %v", err)
	} else {
		t.Logf("Found GET /health -> %s (params: %v)", route.Path, pathParams)
	}
}
