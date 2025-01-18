package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareSetup(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CorsAllowedOrigins: []string{"http://localhost:3000"},
			CorsAllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			CorsAllowedHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
			CorsMaxAge:         3600,
			RequestTimeout:     30 * time.Second,
		},
		RateLimit: config.RateLimitConfig{
			Enabled:    true,
			Rate:       100,
			Burst:      5,
			TimeWindow: time.Minute,
			PerIP:      true,
		},
	}

	// Create mock logger
	mockLogger := logger.NewMockLogger()

	// Create middleware manager
	mw := New(mockLogger, cfg)

	// Create Echo instance
	e := echo.New()

	// Setup middleware
	mw.Setup(e)

	// Verify logger calls
	assert.Len(t, mockLogger.InfoCalls, 1)
	assert.Equal(t, "middleware configuration", mockLogger.InfoCalls[0].Message)
}

func TestRequestIDMiddleware(t *testing.T) {
	// Create test config
	cfg := &config.Config{}

	// Create mock logger
	mockLogger := logger.NewMockLogger()

	// Create middleware manager
	mw := New(mockLogger, cfg)

	// Create Echo instance
	e := echo.New()

	// Add request ID middleware
	e.Use(mw.requestID())

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create test handler
	handler := func(c echo.Context) error {
		requestID := c.Get("request_id")
		assert.NotNil(t, requestID)
		assert.NotEmpty(t, requestID)
		return nil
	}

	// Execute middleware
	h := mw.requestID()(handler)
	err := h(c)

	// Assert no errors
	assert.NoError(t, err)

	// Verify logger calls
	assert.Len(t, mockLogger.DebugCalls, 1)
	assert.Equal(t, "incoming request", mockLogger.DebugCalls[0].Message)
}
