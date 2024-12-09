package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func setupTestServer(t *testing.T) (*echo.Echo, *Manager) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CorsAllowedOrigins:   []string{"https://jonesrussell.github.io"},
			CorsAllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			CorsAllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
			CorsMaxAge:           3600,
			CorsAllowCredentials: true,
			RequestTimeout:       30 * time.Second,
		},
		RateLimit: config.RateLimitConfig{
			Enabled:    true,
			Rate:       5,
			Burst:      3,
			TimeWindow: time.Minute,
			PerIP:      true,
		},
	}

	e := echo.New()
	m := New(logger, cfg)
	m.Setup(e)

	return e, m
}

func TestCORSMiddleware(t *testing.T) {
	e, _ := setupTestServer(t)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		checkHeaders   func(*testing.T, http.Header)
	}{
		{
			name:           "Valid origin",
			origin:         "https://jonesrussell.github.io",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, h http.Header) {
				assert.Equal(t, "https://jonesrussell.github.io", h.Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", h.Get("Access-Control-Allow-Credentials"))
			},
		},
		{
			name:           "OPTIONS request",
			origin:         "https://jonesrussell.github.io",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			checkHeaders: func(t *testing.T, h http.Header) {
				assert.Contains(t, h.Get("Access-Control-Allow-Methods"), "GET")
				assert.Contains(t, h.Get("Access-Control-Allow-Headers"), "Content-Type")
			},
		},
		{
			name:           "Invalid origin",
			origin:         "https://invalid-origin.com",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, h http.Header) {
				assert.Empty(t, h.Get("Access-Control-Allow-Origin"))
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
			tt.checkHeaders(t, rec.Header())
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	e, _ := setupTestServer(t)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	headers := rec.Header()
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Nosniff"))
	assert.Contains(t, headers.Get("Strict-Transport-Security"), "max-age=31536000")
}

func TestRateLimiter(t *testing.T) {
	e, _ := setupTestServer(t)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	for i := 0; i < 8; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if i < 5 {
			assert.Equal(t, http.StatusOK, rec.Code, "Request %d should succeed", i+1)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, rec.Code, "Request %d should be rate limited", i+1)
		}
	}
}

func TestRequestLogger(t *testing.T) {
	// Create a buffer to capture logs
	var buf bytes.Buffer

	// Create a test logger with the buffer as output
	testLogger := zaptest.NewLogger(t).WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		buf.WriteString(entry.Message + "\n")
		return nil
	}))

	// Setup with our test logger
	e := echo.New()
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CorsAllowedOrigins:   []string{"https://jonesrussell.github.io"},
			CorsAllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			CorsAllowedHeaders:   []string{"Origin", "Content-Type", "Accept"},
			CorsMaxAge:           3600,
			CorsAllowCredentials: true,
			RequestTimeout:       30 * time.Second,
		},
		RateLimit: config.RateLimitConfig{
			Enabled: true,
			Rate:    5,
			Burst:   3,
		},
	}

	m := New(testLogger, cfg)
	m.Setup(e)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderXRequestID))

	// Verify logs
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Expected log output")
	assert.Contains(t, logOutput, "incoming request")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "GET")
}

func TestErrorHandler(t *testing.T) {
	e, _ := setupTestServer(t)

	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test error", response["error"])
	assert.Equal(t, float64(http.StatusBadRequest), response["code"])
	assert.NotEmpty(t, response["request_id"])
}

func TestTimeoutMiddleware(t *testing.T) {
	e, _ := setupTestServer(t)

	e.GET("/slow", func(c echo.Context) error {
		time.Sleep(50 * time.Millisecond)
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPanicRecovery(t *testing.T) {
	e, _ := setupTestServer(t)

	e.GET("/panic", func(c echo.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
