package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingMiddleware(t *testing.T) {
	// Create an observer core for testing logs
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	// Create test handler that returns different status codes
	testCases := []struct {
		name       string
		handler    echo.HandlerFunc
		status     int
		expectLogs int
	}{
		{
			name: "successful request",
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			},
			status:     http.StatusOK,
			expectLogs: 1,
		},
		{
			name: "error request",
			handler: func(c echo.Context) error {
				c.Response().Status = http.StatusBadRequest
				return echo.NewHTTPError(http.StatusBadRequest, "bad request")
			},
			status:     http.StatusBadRequest,
			expectLogs: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear logs before each test
			logs.TakeAll()

			// Create Echo instance and context
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Create middleware
			middleware := LoggingMiddleware(logger)
			handler := middleware(tc.handler)

			// Execute request
			err := handler(c)

			// Check error handling
			if tc.status >= 400 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify logs
			logEntries := logs.TakeAll()
			assert.Equal(t, tc.expectLogs, len(logEntries))

			if len(logEntries) > 0 {
				entry := logEntries[0]
				assert.Equal(t, "http request", entry.Message)
				assert.Equal(t, req.Method, entry.Context[0].String)      // method
				assert.Equal(t, req.URL.Path, entry.Context[1].String)    // path
				assert.Equal(t, tc.status, int(entry.Context[2].Integer)) // status
				assert.Equal(t, "192.0.2.1", entry.Context[4].String)     // ip
			}
		})
	}
}

func TestLoggingMiddleware_RealIP(t *testing.T) {
	// Create an observer core for testing logs
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	// Create Echo instance
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create middleware
	middleware := LoggingMiddleware(logger)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Execute request
	_ = handler(c)

	// Verify logs
	logEntries := logs.TakeAll()
	assert.Equal(t, 1, len(logEntries))
	assert.Equal(t, "192.168.1.1", logEntries[0].Context[4].String) // ip
}
