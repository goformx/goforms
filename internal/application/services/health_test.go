package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mocklog "github.com/jonesrussell/goforms/test/mocks/logging"
	"github.com/labstack/echo/v4"
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
				if err == nil {
					t.Error("HandleHealthCheck() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("HandleHealthCheck() error = %v, want nil", err)
				}
			}

			// Verify response
			if rec.Code != tt.wantStatus {
				t.Errorf("HandleHealthCheck() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var gotBody map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &gotBody); err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			// Compare response bodies
			if !deepEqual(t, tt.wantBody, gotBody) {
				t.Errorf("HandleHealthCheck() body = %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}

// deepEqual recursively compares two maps for equality
func deepEqual(t *testing.T, want, got map[string]interface{}) bool {
	if len(want) != len(got) {
		return false
	}
	for key, wantVal := range want {
		gotVal, exists := got[key]
		if !exists {
			return false
		}
		switch v := wantVal.(type) {
		case map[string]interface{}:
			if g, ok := gotVal.(map[string]interface{}); !ok || !deepEqual(t, v, g) {
				return false
			}
		default:
			if wantVal != gotVal {
				return false
			}
		}
	}
	return true
}
