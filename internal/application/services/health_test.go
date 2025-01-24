package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	mocklog "github.com/jonesrussell/goforms/test/mocks/logging"
)

type mockPingContexter struct {
	err error
}

func (m *mockPingContexter) PingContext(ctx echo.Context) error {
	return m.err
}

func TestHealthHandler_HandleHealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		pingError   error
		wantStatus  int
		wantBody    map[string]interface{}
		wantLogCall bool
	}{
		{
			name:       "healthy service",
			pingError:  nil,
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"status": "healthy",
				},
			},
			wantLogCall: false,
		},
		{
			name:       "unhealthy service",
			pingError:  errors.New("db connection failed"),
			wantStatus: http.StatusInternalServerError,
			wantBody: map[string]interface{}{
				"success": false,
				"error":   "Service is not healthy",
			},
			wantLogCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockLogger := mocklog.NewMockLogger()
			mockDB := &mockPingContexter{err: tt.pingError}
			handler := NewHealthHandler(mockLogger, mockDB)

			// Create request and recorder
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			// Execute
			err := handler.HandleHealthCheck(c)

			// Assert
			if tt.wantLogCall {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify response
			assert.Equal(t, tt.wantStatus, rec.Code)
			var gotBody map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &gotBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantBody, gotBody)
		})
	}
}
