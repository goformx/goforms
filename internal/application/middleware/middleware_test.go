package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestMiddlewareSetup(t *testing.T) {
	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("middleware configuration")

	// Create middleware manager
	mw := New(mockLogger)

	// Create Echo instance
	e := echo.New()

	// Setup middleware
	mw.Setup(e)

	// Verify logger calls
	assert.NoError(t, mockLogger.Verify(), "logger expectations not met")
}

func TestRequestIDMiddleware(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("incoming request",
		logging.String("request_id", mock.Anything),
		logging.String("method", "GET"),
		logging.String("path", "/"),
		logging.String("remote_addr", "192.0.2.1:1234"),
	)

	e := echo.New()
	m := New(mockLogger)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Add middleware
	e.Use(m.requestID())

	// Create test handler
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Process the request
	e.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, mockLogger.Verify(), "logger expectations not met")
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mockLogger := mocklogging.NewMockLogger()

	// Create middleware manager
	m := New(mockLogger)

	// Create handler with security headers middleware
	handler := m.securityHeaders()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Execute
	err := handler(c)
	assert.NoError(t, err)

	// Assert headers
	headers := rec.Header()
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", headers.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	assert.Contains(t, headers.Get("Content-Security-Policy"), "default-src 'self'")
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
	assert.Equal(t, "same-origin", headers.Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "require-corp", headers.Get("Cross-Origin-Embedder-Policy"))
	assert.Equal(t, "same-origin", headers.Get("Cross-Origin-Resource-Policy"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", headers.Get("Permissions-Policy"))
}
