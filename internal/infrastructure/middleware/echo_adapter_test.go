package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/middleware"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEchoAdapter_ToEchoMiddleware tests the Echo adapter conversion
func TestEchoAdapter_ToEchoMiddleware(t *testing.T) {
	// Create a simple test middleware
	testMiddleware := &testMiddleware{
		name: "test-middleware",
	}

	// Create Echo adapter
	adapter := middleware.NewEchoAdapter(testMiddleware)
	echoMiddleware := adapter.ToEchoMiddleware()

	// Create Echo instance
	e := echo.New()

	// Add test handler
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test response")
	})

	// Apply our middleware
	e.Use(echoMiddleware)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()

	// Serve request
	e.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test response")
	assert.Equal(t, "test-header-value", rec.Header().Get("X-Test-Header"))
}

// TestEchoRequest tests the Echo request wrapper
func TestEchoRequest(t *testing.T) {
	// Create Echo instance
	e := echo.New()

	// Create test request
	req := httptest.NewRequest(http.MethodPost, "/test?param=value", strings.NewReader("test body"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("User-Agent", "test-agent")

	// Create Echo context
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create our request wrapper
	echoReq := middleware.NewEchoRequest(c)

	// Test basic properties
	assert.Equal(t, http.MethodPost, echoReq.Method())
	assert.Equal(t, "/test", echoReq.Path())
	assert.Equal(t, "value", echoReq.Query()["param"][0])
	assert.Equal(t, "text/plain", echoReq.ContentType())
	assert.Equal(t, "test-agent", echoReq.UserAgent())
	assert.NotNil(t, echoReq.Body())
}

// TestEchoResponseWrapper tests the Echo response wrapper
func TestEchoResponseWrapper(t *testing.T) {
	// Create Echo instance
	e := echo.New()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create our response
	resp := core.NewResponse(http.StatusOK)
	resp.SetContentType("application/json")
	resp.SetBodyBytes([]byte(`{"message":"test"}`))

	// Create response wrapper
	wrapper := middleware.NewEchoResponseWrapper(resp, c)

	// Apply response
	err := wrapper.ApplyToEcho()
	require.NoError(t, err)

	// Verify response was applied
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.Equal(t, "application/json", c.Response().Header().Get("Content-Type"))
}

// TestEchoResponseBuilder tests the Echo response builder
func TestEchoResponseBuilder(t *testing.T) {
	// Create Echo instance
	e := echo.New()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create response builder
	builder := middleware.NewEchoResponseBuilder(c)

	// Build response
	resp := builder.
		Status(http.StatusCreated).
		ContentType("application/json").
		JSON(map[string]string{"message": "created"}).
		Header("X-Custom", "value").
		Build()

	// Verify response
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
	assert.Equal(t, "application/json", resp.ContentType())
	assert.JSONEq(t, `{"message":"created"}`, string(resp.BodyBytes()))
	assert.Equal(t, "value", resp.Headers().Get("X-Custom"))
}

// TestEchoResponseWriter tests the Echo response writer
func TestEchoResponseWriter(t *testing.T) {
	// Create Echo instance
	e := echo.New()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create response writer
	writer := middleware.NewEchoResponseWriter(c)

	// Write some content
	_, err := writer.WriteString("Hello, World!")
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, "Hello, World!", writer.String())
	assert.Equal(t, []byte("Hello, World!"), writer.Bytes())
}

// testMiddleware is a simple test middleware for testing
type testMiddleware struct {
	name string
}

func (m *testMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Add a test header to the response
	resp := next(ctx, req)
	resp.AddHeader("X-Test-Header", "test-header-value")

	return resp
}

func (m *testMiddleware) Name() string {
	return m.name
}

func (m *testMiddleware) Priority() int {
	return 0
}
