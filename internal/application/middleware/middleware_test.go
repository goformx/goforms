package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

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
	mockLogger.AssertExpectations(t)
}

func TestRequestIDMiddleware(t *testing.T) {
	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("incoming request",
		logging.String("request_id", "test-id"),
	)

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
	mockLogger.AssertExpectations(t)
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("adding security headers")

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
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))

	// Verify logger calls
	mockLogger.AssertExpectations(t)
}
