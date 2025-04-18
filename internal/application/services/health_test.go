package services_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mocklog "github.com/jonesrussell/goforms/test/mocks/logging"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/goforms/internal/application/services"
)

// MockDB is a mock implementation of PingContexter
type MockDB struct {
	mock.Mock
}

func (m *MockDB) PingContext(ctx echo.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// mockPingContexter is a mock implementation of PingContexter
type mockPingContexter struct {
	err error
}

func (m *mockPingContexter) PingContext(ctx echo.Context) error {
	return m.err
}

func TestHealthHandler_HandleHealthCheck(t *testing.T) {
	// Setup
	mockLogger := &mocklog.MockLogger{}
	mockDB := &MockDB{}
	mockDB.On("PingContext", mock.Anything).Return(nil)

	handler := services.NewHealthHandler(mockLogger, mockDB)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	execErr := handler.HandleHealthCheck(c)
	require.NoError(t, execErr)

	// Parse response
	var gotBody map[string]any
	if unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &gotBody); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal response: %v", unmarshalErr)
	}

	// Verify response
	wantBody := map[string]any{
		"status": "healthy",
	}
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, wantBody, gotBody)

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name      string
		wantCode  int
		wantBody  map[string]any
		wantError bool
	}{
		{
			name:     "healthy",
			wantCode: 200,
			wantBody: map[string]any{
				"status": "ok",
				"data": map[string]any{
					"uptime": time.Duration(0),
				},
			},
			wantError: false,
		},
		{
			name:     "unhealthy",
			wantCode: 503,
			wantBody: map[string]any{
				"status": "error",
				"error":  "service unavailable",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(echo.GET, "/health", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockLogger := mocklog.NewMockLogger()
			mockDB := &mockPingContexter{err: nil}
			if tt.wantError {
				mockDB.err = errors.New("service unavailable")
			}

			h := services.NewHealthHandler(mockLogger, mockDB)
			err := h.HandleHealthCheck(c)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			var gotBody map[string]any
			err = json.Unmarshal(rec.Body.Bytes(), &gotBody)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, gotBody)
		})
	}
}

func TestHealthHandler_HandleHealthCheck_New(t *testing.T) {
	// Setup
	mockLogger := &mocklog.MockLogger{}
	mockDB := &mockPingContexter{err: nil}
	handler := services.NewHealthHandler(mockLogger, mockDB)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	execErr := handler.HandleHealthCheck(c)
	require.NoError(t, execErr)

	// Parse response
	var gotBody map[string]any
	if unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &gotBody); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal response: %v", unmarshalErr)
	}

	// Verify response
	wantBody := map[string]any{
		"status": "ok",
		"time":   gotBody["time"], // Use actual time from response
	}
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, wantBody, gotBody)
}
