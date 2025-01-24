package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

func TestLoggingMiddleware(t *testing.T) {
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
			// Create a new mock logger for each test
			mockLogger := logging.NewMockLogger()

			// Create Echo instance and context
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Create middleware
			middleware := LoggingMiddleware(mockLogger)
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
			infoCalls := mockLogger.InfoCalls
			assert.Equal(t, tc.expectLogs, len(infoCalls))

			if len(infoCalls) > 0 {
				call := infoCalls[0]
				assert.Equal(t, "http request", call.Message)
				assert.Contains(t, call.Fields, logging.String("method", req.Method))
				assert.Contains(t, call.Fields, logging.String("path", req.URL.Path))
				assert.Contains(t, call.Fields, logging.Int("status", tc.status))
				assert.Contains(t, call.Fields, logging.String("ip", "192.0.2.1"))
			}
		})
	}
}

func TestLoggingMiddleware_RealIP(t *testing.T) {
	// Create a mock logger for testing
	mockLogger := logging.NewMockLogger()

	// Create Echo instance
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create middleware
	middleware := LoggingMiddleware(mockLogger)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Execute request
	_ = handler(c)

	// Verify logs
	infoCalls := mockLogger.InfoCalls
	assert.Equal(t, 1, len(infoCalls))
	assert.Contains(t, infoCalls[0].Fields, logging.String("ip", "192.168.1.1"))
}
