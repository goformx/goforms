package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/application/middleware"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestLoggingMiddleware(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	loggingMiddleware := middleware.LoggingMiddleware(mockLogger)

	handler := func(c echo.Context) error {
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	mockLogger.ExpectInfo("request completed").WithFields(map[string]any{
		"method":   http.MethodGet,
		"path":     "/test",
		"status":   http.StatusOK,
		"duration": mock.Anything,
	})

	if err := loggingMiddleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mockLogger.Verify(); err != nil {
		t.Fatalf("logger expectations not met: %v", err)
	}
}

func TestLoggingMiddlewareWithError(t *testing.T) {
	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("http request").WithFields(map[string]any{
		"method":      "GET",
		"path":        "/",
		"status":      500,
		"latency":     mocklogging.AnyValue{},
		"remote_addr": mocklogging.AnyValue{},
		"user_agent":  "",
		"error":       "test error",
	})

	// Create Echo instance
	e := echo.New()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create handler that returns an error
	handler := func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	}

	// Create middleware and wrap handler
	h := middleware.LoggingMiddleware(mockLogger)(handler)

	// Test middleware
	_ = h(c)

	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}

func TestLoggingMiddlewareWithPanic(t *testing.T) {
	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("http request").WithFields(map[string]any{
		"method":      "GET",
		"path":        "/",
		"status":      500,
		"latency":     mocklogging.AnyValue{},
		"remote_addr": mocklogging.AnyValue{},
		"user_agent":  "",
		"panic":       "test panic",
	})

	// Create Echo instance
	e := echo.New()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create handler that panics
	handler := func(c echo.Context) error {
		panic("test panic")
	}

	// Create middleware and wrap handler
	h := middleware.LoggingMiddleware(mockLogger)(handler)

	// Test middleware
	_ = h(c)

	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}

func TestLoggingMiddleware_RealIP(t *testing.T) {
	// Create a mock logger for testing
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("http request").WithFields(map[string]any{
		"method":      "GET",
		"path":        "/test",
		"status":      200,
		"latency":     mocklogging.AnyValue{},
		"remote_addr": mocklogging.AnyValue{},
		"user_agent":  "",
	})

	// Create Echo instance
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create middleware
	mw := middleware.LoggingMiddleware(mockLogger)
	handler := mw(func(c echo.Context) error {
		c.Response().WriteHeader(http.StatusOK)
		return c.String(http.StatusOK, "success")
	})

	// Execute request
	if err := handler(c); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify logs
	if verifyErr := mockLogger.Verify(); verifyErr != nil {
		t.Errorf("logger expectations not met: %v", verifyErr)
	}
}
