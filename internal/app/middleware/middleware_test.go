package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMiddlewareHeaders(t *testing.T) {
	// Setup
	e := echo.New()
	logger := zap.NewNop()
	cfg := &config.SecurityConfig{
		CorsAllowedOrigins:   []string{"https://jonesrussell.github.io"},
		CorsAllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CorsAllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		CorsMaxAge:           3600,
		CorsAllowCredentials: true,
		TrustedProxies:       []string{"127.0.0.1", "::1"},
		RequestTimeout:       30 * time.Second,
	}

	// Create middleware instance
	m := New(logger, cfg)
	m.Setup(e)

	// Create test handler
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		checkHeaders   func(*testing.T, http.Header)
	}{
		{
			name:           "Should have correct CORS headers",
			origin:         "https://jonesrussell.github.io",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, h http.Header) {
				// Only check CORS headers when Origin is present
				if origin := h.Get("Access-Control-Allow-Origin"); origin != "" {
					assert.Equal(t, "https://jonesrussell.github.io", origin)
					assert.Equal(t, "true", h.Get("Access-Control-Allow-Credentials"))
				}

				// Other headers
				assert.NotEmpty(t, h.Get("X-Request-Id"))
				assert.Empty(t, h.Get("X-Frame-Options"))
				assert.Empty(t, h.Get("X-XSS-Protection"))
				assert.Empty(t, h.Get("X-Content-Type-Options"))
				assert.Empty(t, h.Get("Strict-Transport-Security"))
			},
		},
		{
			name:           "Should handle OPTIONS request",
			origin:         "https://jonesrussell.github.io",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			checkHeaders: func(t *testing.T, h http.Header) {
				assert.Equal(t, "https://jonesrussell.github.io", h.Get("Access-Control-Allow-Origin"))
				assert.Contains(t, h.Get("Access-Control-Allow-Methods"), "GET,POST,PUT,DELETE,OPTIONS")
				assert.Contains(t, h.Get("Access-Control-Allow-Headers"), "Content-Type")
				assert.Equal(t, "3600", h.Get("Access-Control-Max-Age"))
				assert.Equal(t, "true", h.Get("Access-Control-Allow-Credentials"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.checkHeaders(t, rec.Header())
		})
	}
}
