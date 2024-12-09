package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCORSConfiguration(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	e := echo.New()
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CorsAllowedOrigins: []string{"https://jonesrussell.github.io"},
			CorsAllowedMethods: []string{"GET", "POST", "OPTIONS"},
			CorsAllowedHeaders: []string{"Origin", "Content-Type", "Accept"},
			CorsMaxAge:         3600,
		},
	}

	app := &App{
		logger: logger,
		echo:   e,
		config: cfg,
	}

	app.setupMiddleware()

	tests := []struct {
		name            string
		method          string
		origin          string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "Valid Origin - GET Request",
			method:         "GET",
			origin:         "https://jonesrussell.github.io",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://jonesrussell.github.io",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:           "Invalid Origin - GET Request",
			method:         "GET",
			origin:         "https://invalid-origin.com",
			expectedStatus: http.StatusForbidden,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "",
			},
		},
		{
			name:           "OPTIONS Preflight - Valid Origin",
			method:         "OPTIONS",
			origin:         "https://jonesrussell.github.io",
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://jonesrussell.github.io",
				"Access-Control-Allow-Methods":     "GET,POST,OPTIONS",
				"Access-Control-Allow-Headers":     "Origin,Content-Type,Accept",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "3600",
			},
		},
		{
			name:           "OPTIONS Preflight - Invalid Origin",
			method:         "OPTIONS",
			origin:         "https://invalid-origin.com",
			expectedStatus: http.StatusForbidden,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "",
			},
		},
	}

	// Test endpoint
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set(echo.HeaderOrigin, tt.origin)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code, "Status code mismatch")

			// Check headers
			for key, expectedValue := range tt.expectedHeaders {
				actualValue := rec.Header().Get(key)
				assert.Equal(t, expectedValue, actualValue,
					"Header mismatch for %s", key)
			}

			// Additional logging for debugging
			if t.Failed() {
				t.Logf("Response Headers: %v", rec.Header())
				t.Logf("Response Body: %s", rec.Body.String())
			}
		})
	}
}

func TestCORSWithRequestContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	e := echo.New()
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CorsAllowedOrigins: []string{"https://jonesrussell.github.io"},
			CorsAllowedMethods: []string{"GET", "POST", "OPTIONS"},
			CorsAllowedHeaders: []string{"Origin", "Content-Type", "Accept"},
			CorsMaxAge:         3600,
		},
	}

	app := &App{
		logger: logger,
		echo:   e,
		config: cfg,
	}

	app.setupMiddleware()

	// Test endpoint that requires context
	e.POST("/api/subscriptions", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, map[string]string{"status": "success"})
	})

	t.Run("POST Request with Context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", nil)
		req.Header.Set(echo.HeaderOrigin, "https://jonesrussell.github.io")
		req.Header.Set(echo.HeaderContentType, "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "https://jonesrussell.github.io",
			rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true",
			rec.Header().Get("Access-Control-Allow-Credentials"))
	})
}
