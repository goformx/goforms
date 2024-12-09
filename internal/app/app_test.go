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

func TestCORSMiddleware(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	e := echo.New()
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CorsAllowedOrigins: []string{"https://jonesrussell.github.io"},
			CorsAllowedMethods: []string{"GET", "POST", "OPTIONS"},
			CorsAllowedHeaders: []string{"Origin", "Content-Type"},
			CorsMaxAge:         3600,
		},
		RateLimit: config.RateLimitConfig{
			Rate: 100, // Higher rate for testing
		},
	}

	app := &App{
		logger: logger,
		echo:   e,
		config: cfg,
	}

	app.setupMiddleware()

	// Test endpoint
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	tests := []struct {
		name            string
		method          string
		origin          string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "allowed_origin",
			method:         "GET",
			origin:         "https://jonesrussell.github.io",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://jonesrussell.github.io",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:           "preflight_request",
			method:         "OPTIONS",
			origin:         "https://jonesrussell.github.io",
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://jonesrussell.github.io",
				"Access-Control-Allow-Methods":     "GET,POST,OPTIONS",
				"Access-Control-Allow-Headers":     "Origin,Content-Type",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "3600",
			},
		},
		{
			name:           "preflight_request_invalid_origin",
			method:         "OPTIONS",
			origin:         "https://invalid-origin.com",
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Allow": "OPTIONS, GET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set(echo.HeaderOrigin, tt.origin)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedHeaders != nil {
				for key, expectedValue := range tt.expectedHeaders {
					assert.Equal(t, expectedValue, rec.Header().Get(key))
				}
			}
		})
	}
}
