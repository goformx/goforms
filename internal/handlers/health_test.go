package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type MockPingContexter struct {
	err error
}

func (m *MockPingContexter) PingContext(c echo.Context) error {
	return m.err
}

func TestHealthHandler(t *testing.T) {
	t.Run("successful health check", func(t *testing.T) {
		// Setup
		e := echo.New()
		mockLogger := logger.NewMockLogger()
		mockDB := &MockPingContexter{}
		handler := NewHealthHandler(mockLogger, mockDB)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Test
		err := handler.HandleHealth(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify logger calls
		assert.True(t, len(mockLogger.DebugCalls) > 0)
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		e := echo.New()
		mockLogger := logger.NewMockLogger()
		mockDB := &MockPingContexter{err: errors.New("database error")}
		handler := NewHealthHandler(mockLogger, mockDB)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Test
		err := handler.HandleHealth(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		// Verify logger calls
		assert.True(t, len(mockLogger.ErrorCalls) > 0)
	})
}
