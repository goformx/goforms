package openapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/goformx/goforms/internal/application/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data - minimal OpenAPI spec for testing
const testOpenAPISpec = `
openapi: 3.0.0
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
  /test/{id}:
    get:
      summary: Test endpoint with path parameter
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  message:
                    type: string
  /auth:
    get:
      summary: Auth endpoint
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
                  authenticated:
                    type: boolean
components:
  securitySchemes:
    SessionAuth:
      type: apiKey
      in: header
      name: X-Session-Token
`

func createTestRouter(t *testing.T) routers.Router {
	t.Helper()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(testOpenAPISpec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	return router
}

func TestNewOpenAPIRequestValidator(t *testing.T) {
	router := createTestRouter(t)
	validator := openapi.NewOpenAPIRequestValidator(router)

	assert.NotNil(t, validator)
	assert.IsType(t, &openapi.OpenAPIRequestValidator{}, validator)
}

func TestOpenAPIRequestValidator_ValidateRequest_Success(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	// Test valid request
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	route, pathParams, err := validator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateRequest(req, route, pathParams)
	require.NoError(t, err)
}

func TestOpenAPIRequestValidator_ValidateRequest_WithPathParams(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	// Test valid request with path parameters
	req := httptest.NewRequest(http.MethodGet, "/test/123", http.NoBody)
	route, pathParams, err := validator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateRequest(req, route, pathParams)
	require.NoError(t, err)
	assert.Equal(t, "123", pathParams["id"])
}

func TestOpenAPIRequestValidator_ValidateRequest_WithAuth(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	// Test valid request with authentication
	req := httptest.NewRequest(http.MethodGet, "/auth", http.NoBody)
	req.Header.Set("X-Session-Token", "test-token")
	route, pathParams, err := validator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateRequest(req, route, pathParams)
	require.NoError(t, err)
}

func TestOpenAPIRequestValidator_ValidateRequest_InvalidPath(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	// Test invalid path
	req := httptest.NewRequest(http.MethodGet, "/invalid", http.NoBody)
	route, pathParams, err := validator.FindRoute(req)
	require.Error(t, err)
	assert.Nil(t, route)
	assert.Nil(t, pathParams)
}

func TestOpenAPIRequestValidator_FindRoute_Success(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	route, pathParams, err := validator.FindRoute(req)

	require.NoError(t, err)
	assert.NotNil(t, route)
	assert.NotNil(t, pathParams)
}

func TestOpenAPIRequestValidator_FindRoute_NotFound(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", http.NoBody)
	route, pathParams, err := validator.FindRoute(req)

	require.Error(t, err)
	assert.Nil(t, route)
	assert.Nil(t, pathParams)
	assert.Contains(t, err.Error(), "failed to find route")
}

func TestNewOpenAPIResponseValidator(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIResponseValidator(router).(*openapi.OpenAPIResponseValidator)
	require.True(t, ok)

	assert.NotNil(t, validator)
	assert.IsType(t, &openapi.OpenAPIResponseValidator{}, validator)
}

func TestOpenAPIResponseValidator_ValidateResponse_Success(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIResponseValidator(router).(*openapi.OpenAPIResponseValidator)
	require.True(t, ok)

	// Create test request and response
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}
	body := []byte(`{"message": "test"}`)

	// Get route for validation
	requestValidator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	route, pathParams, err := requestValidator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateResponse(req, resp, body, route, pathParams)
	require.NoError(t, err)
}

func TestOpenAPIResponseValidator_ValidateResponse_InvalidStatus(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIResponseValidator(router).(*openapi.OpenAPIResponseValidator)
	require.True(t, ok)

	// Create test request and response with invalid status
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	resp := &http.Response{
		StatusCode: http.StatusNotFound, // Not defined in spec
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}
	body := []byte(`{"error": "not found"}`)

	// Get route for validation
	requestValidator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	route, pathParams, err := requestValidator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateResponse(req, resp, body, route, pathParams)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "response validation failed")
}

func TestOpenAPIResponseValidator_ValidateResponse_InvalidContentType(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIResponseValidator(router).(*openapi.OpenAPIResponseValidator)
	require.True(t, ok)

	// Create test request and response with invalid content type
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"text/plain"}, // Not defined in spec
		},
	}
	body := []byte("plain text response")

	// Get route for validation
	requestValidator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	route, pathParams, err := requestValidator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateResponse(req, resp, body, route, pathParams)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "response validation failed")
}

func TestOpenAPIResponseValidator_ValidateResponse_EmptyBody(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIResponseValidator(router).(*openapi.OpenAPIResponseValidator)
	require.True(t, ok)

	// Create test request and response with empty body
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}
	body := []byte{}

	// Get route for validation
	requestValidator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	route, pathParams, err := requestValidator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateResponse(req, resp, body, route, pathParams)
	require.NoError(t, err) // Empty body should be valid for JSON
}

func TestOpenAPIResponseValidator_ValidateResponse_NilRoute(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIResponseValidator(router).(*openapi.OpenAPIResponseValidator)
	require.True(t, ok)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
	}
	body := []byte(`{"message": "test"}`)

	// Test with nil route
	err := validator.ValidateResponse(req, resp, body, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "response validation failed")
}

func TestOpenAPIRequestValidator_ValidateRequest_NilRoute(t *testing.T) {
	router := createTestRouter(t)
	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

	// Test with nil route
	err := validator.ValidateRequest(req, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request validation failed")
}

func TestOpenAPIRequestValidator_ValidateRequest_WithBody(t *testing.T) {
	// Create a spec with request body validation
	const specWithBody = `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    post:
      summary: Test endpoint with body
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
                age:
                  type: integer
      responses:
        '200':
          description: OK
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(specWithBody))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	// Test valid request with body
	body := strings.NewReader(`{"name": "test", "age": 25}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")

	route, pathParams, err := validator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateRequest(req, route, pathParams)
	require.NoError(t, err)
}

func TestOpenAPIRequestValidator_ValidateRequest_InvalidBody(t *testing.T) {
	// Create a spec with request body validation
	const specWithBody = `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    post:
      summary: Test endpoint with body
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
      responses:
        '200':
          description: OK
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(specWithBody))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	validator, ok := openapi.NewOpenAPIRequestValidator(router).(*openapi.OpenAPIRequestValidator)
	require.True(t, ok)

	// Test invalid request with missing required field
	body := strings.NewReader(`{"age": 25}`) // Missing required 'name' field
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")

	route, pathParams, err := validator.FindRoute(req)
	require.NoError(t, err)

	err = validator.ValidateRequest(req, route, pathParams)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request validation failed")
}
