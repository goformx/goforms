package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/test/mocks"
)

func TestMiddlewareSetup(t *testing.T) {
	// Create mock logger
	mockLogger := mocks.NewLogger()

	// Create middleware manager
	mw := New(mockLogger)

	// Create Echo instance
	e := echo.New()

	// Setup middleware
	mw.Setup(e)

	// Verify logger calls
	assert.Len(t, mockLogger.InfoCalls, 1)
	assert.Equal(t, "middleware configuration", mockLogger.InfoCalls[0].Message)
}

func TestRequestIDMiddleware(t *testing.T) {
	// Create mock logger
	mockLogger := mocks.NewLogger()

	// Create middleware manager
	mw := New(mockLogger)

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

func TestSecurityHeadersMiddleware(t *testing.T) {
	// Create mock logger
	mockLogger := mocks.NewLogger()

	// Create middleware manager
	mw := New(mockLogger)

	// Create Echo instance
	e := echo.New()

	// Add security headers middleware
	e.Use(mw.securityHeaders())

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create test handler
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	// Execute middleware
	h := mw.securityHeaders()(handler)
	err := h(c)

	// Assert no errors
	assert.NoError(t, err)

	// Verify security headers
	headers := rec.Header()
	assert.NotEmpty(t, headers.Get("Content-Security-Policy"))
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
	assert.Equal(t, "same-origin", headers.Get("Cross-Origin-Resource-Policy"))
	assert.Empty(t, headers.Get("X-XSS-Protection"))
	assert.Empty(t, headers.Get("Server"))
}
