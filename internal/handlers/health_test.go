package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockDB struct {
	PingError error
}

func (m *MockDB) PingContext(_ context.Context) error {
	return m.PingError
}

func setupHealthTest() (*echo.Echo, *HealthHandler, *MockDB) {
	e := echo.New()
	logger, _ := zap.NewDevelopment()
	mockDB := &MockDB{}
	handler := NewHealthHandler(mockDB, logger)
	return e, handler, mockDB
}

func TestHealthCheck(t *testing.T) {
	e, handler, mockDB := setupHealthTest()
	handler.Register(e)

	t.Run("successful health check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Check(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "ok", response["status"])
		assert.Equal(t, "ok", response["db_status"])
		assert.NotEmpty(t, response["timestamp"])
	})

	t.Run("database error", func(t *testing.T) {
		mockDB.PingError = errors.New("database connection error")

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Check(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		var response map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		assert.Equal(t, "degraded", response["status"])
		assert.Equal(t, "error", response["db_status"])
		assert.NotEmpty(t, response["timestamp"])
	})
}
