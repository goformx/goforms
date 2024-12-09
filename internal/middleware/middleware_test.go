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
	t.Skip("Security headers are handled by Nginx")
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
	var buf bytes.Buffer

	// Create a test logger that writes to our buffer
	testLogger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(&buf),
			zapcore.DebugLevel,
		)
	})))

	e := echo.New()
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CorsAllowedOrigins:   []string{"https://jonesrussell.github.io"},
			CorsAllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			CorsAllowedHeaders:   []string{"Origin", "Content-Type", "Accept"},
			CorsMaxAge:           3600,
			CorsAllowCredentials: true,
		},
	}

	mw := New(testLogger, cfg)
	mw.Setup(e)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	logs := buf.String()
	assert.Contains(t, logs, "/test")
	assert.Contains(t, logs, "GET")
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
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)

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
