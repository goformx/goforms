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
	}

	app := &App{
		logger: logger,
		echo:   e,
		config: cfg,
	}

	app.setupMiddleware()

	// Add a test handler
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "allowed origin",
			origin:         "https://jonesrussell.github.io",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedHeader: "https://jonesrussell.github.io",
		},
		{
			name:           "disallowed origin",
			origin:         "https://evil.com",
			method:         "GET",
			expectedStatus: http.StatusForbidden,
			expectedHeader: "",
		},
		{
			name:           "preflight request",
			origin:         "https://jonesrussell.github.io",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			expectedHeader: "https://jonesrussell.github.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set(echo.HeaderOrigin, tt.origin)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, rec.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}
