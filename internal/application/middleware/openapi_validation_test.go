package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/config"
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
