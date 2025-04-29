package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/application/middleware"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	"github.com/jonesrussell/goforms/test/utils"
)

func TestMiddleware_Setup(t *testing.T) {
	ts := utils.NewTestSetup(t)
	defer ts.Close()

	// Create middleware manager
	mw := middleware.New(&middleware.ManagerConfig{
		Logger: ts.Logger,
	})

	// Setup middleware
	mw.Setup(ts.Echo)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := ts.Echo.NewContext(req, rec)

	// Create handler
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Test middleware chain
	err := handler(c)
	ts.RequireNoError(err)

	// Assert security headers
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
	assert.Equal(t, "default-src 'self'", rec.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	ts := utils.NewTestSetup(t)
	defer ts.Close()

	// Create middleware manager
	mw := middleware.New(&middleware.ManagerConfig{
		Logger: ts.Logger,
	})

	// Setup middleware
	mw.Setup(ts.Echo)

	// Create test handler
	ts.Echo.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()

	// Process the request
	ts.Echo.ServeHTTP(rec, req)

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
		assert.Equal(t, expected, got, "header %s mismatch", header)
	}

	// Check CSP header separately as it contains a dynamic nonce
	csp := rec.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "default-src 'self'", "CSP header should contain default-src 'self'")
}

func TestRequestIDMiddleware(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("creating new middleware manager")
	mockLogger.ExpectDebug("processing request ID middleware").WithFields(map[string]any{
		"request_id":  mocklogging.AnyValue{},
		"method":      "GET",
		"path":        "/",
		"remote_addr": mocklogging.AnyValue{},
	})
	mockLogger.ExpectDebug("request ID middleware complete").WithFields(map[string]any{
		"request_id": mocklogging.AnyValue{},
	})

	e := echo.New()
	m := middleware.New(&middleware.ManagerConfig{
		Logger: mockLogger,
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()

	// Setup middleware
	m.Setup(e)

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

func TestRequestID(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("processing request ID middleware").WithFields(map[string]any{
		"request_id": "test-request-id",
		"method":     "GET",
		"path":       "/test",
	})
	mockLogger.ExpectDebug("request ID middleware complete").WithFields(map[string]any{
		"request_id": "test-request-id",
	})

	e := echo.New()
	m := middleware.New(&middleware.ManagerConfig{
		Logger: mockLogger,
	})

	// Create test request with custom request ID
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set(echo.HeaderXRequestID, "test-request-id")
	rec := httptest.NewRecorder()

	// Setup middleware
	m.Setup(e)

	// Create test handler
	e.GET("/test", func(c echo.Context) error {
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

func TestSecurityHeaders(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectDebug("processing security headers").WithFields(map[string]any{
		"path":   "/test",
		"method": "GET",
	})

	csp := "default-src 'self'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"script-src 'self' 'nonce-test'; " +
		"img-src 'self' data:; " +
		"font-src 'self'; " +
		"connect-src 'self'; " +
		"base-uri 'self'; " +
		"form-action 'self'"

	mockLogger.ExpectDebug("built CSP directives").WithFields(map[string]any{
		"csp": csp,
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "Content-Security-Policy",
		"value":  csp,
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "X-Content-Type-Options",
		"value":  "nosniff",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "X-Frame-Options",
		"value":  "SAMEORIGIN",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "X-XSS-Protection",
		"value":  "1; mode=block",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "Referrer-Policy",
		"value":  "strict-origin-when-cross-origin",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "Permissions-Policy",
		"value":  "geolocation=(), microphone=(), camera=()",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "Cross-Origin-Opener-Policy",
		"value":  "same-origin",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "Cross-Origin-Embedder-Policy",
		"value":  "require-corp",
	})
	mockLogger.ExpectDebug("set security header").WithFields(map[string]any{
		"header": "Cross-Origin-Resource-Policy",
		"value":  "same-origin",
	})
	mockLogger.ExpectDebug("removed Server header")
	mockLogger.ExpectDebug("security headers processing complete")

	e := echo.New()
	m := middleware.New(&middleware.ManagerConfig{
		Logger: mockLogger,
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()

	// Setup middleware
	m.Setup(e)

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
	csp = rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("expected CSP to contain default-src 'self', got %q", csp)
	}

	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}
