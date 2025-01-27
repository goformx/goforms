package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestMiddlewareSetup(t *testing.T) {
	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("creating new middleware manager")
	mockLogger.ExpectDebug("setting up middleware")
	mockLogger.ExpectDebug("adding security headers middleware")
	mockLogger.ExpectDebug("adding request ID middleware")
	mockLogger.ExpectDebug("middleware setup complete")

	// Create middleware manager
	mw := New(mockLogger)

	// Create Echo instance
	e := echo.New()

	// Setup middleware
	mw.Setup(e)

	// Verify logger calls
	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("creating new middleware manager")
	mockLogger.ExpectDebug("processing request ID middleware").WithFields(map[string]interface{}{
		"request_id":  mocklogging.AnyValue{},
		"method":      "GET",
		"path":        "/",
		"remote_addr": mocklogging.AnyValue{},
	})
	mockLogger.ExpectDebug("request ID middleware complete").WithFields(map[string]interface{}{
		"request_id": mocklogging.AnyValue{},
	})

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
	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("creating new middleware manager")
	mockLogger.ExpectDebug("processing security headers").WithFields(map[string]interface{}{
		"path":   "/",
		"method": "GET",
	})
	mockLogger.ExpectDebug("generated nonce for request")
	mockLogger.ExpectDebug("added nonce to request context")
	mockLogger.ExpectDebug("built CSP directives").WithFields(map[string]interface{}{
		"csp": mocklogging.AnyValue{},
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "Content-Security-Policy",
		"value":  mocklogging.AnyValue{},
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "X-Content-Type-Options",
		"value":  "nosniff",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "X-Frame-Options",
		"value":  "SAMEORIGIN",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "X-XSS-Protection",
		"value":  "1; mode=block",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "Referrer-Policy",
		"value":  "strict-origin-when-cross-origin",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "Permissions-Policy",
		"value":  "geolocation=(), microphone=(), camera=()",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "Cross-Origin-Opener-Policy",
		"value":  "same-origin",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "Cross-Origin-Embedder-Policy",
		"value":  "require-corp",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]interface{}{
		"header": "Cross-Origin-Resource-Policy",
		"value":  "same-origin",
	})
	mockLogger.ExpectDebug("removed Server header")
	mockLogger.ExpectDebug("security headers processing complete")

	e := echo.New()
	m := New(mockLogger)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Add middleware
	e.Use(m.securityHeaders())

	// Create test handler
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Process the request
	e.ServeHTTP(rec, req)

	// Check security headers
	expectedHeaders := map[string]string{
		"X-Content-Type-Options":       "nosniff",
		"X-Frame-Options":              "SAMEORIGIN",
		"X-XSS-Protection":             "1; mode=block",
		"Referrer-Policy":              "strict-origin-when-cross-origin",
		"Permissions-Policy":           "geolocation=(), microphone=(), camera=()",
		"Cross-Origin-Opener-Policy":   "same-origin",
		"Cross-Origin-Embedder-Policy": "require-corp",
		"Cross-Origin-Resource-Policy": "same-origin",
	}

	for header, expected := range expectedHeaders {
		got := rec.Header().Get(header)
		if got != expected {
			t.Errorf("expected %s header to be %q, got %q", header, expected, got)
		}
	}

	// Check CSP header separately as it contains a dynamic nonce
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("expected CSP to contain default-src 'self', got %q", csp)
	}

	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}
