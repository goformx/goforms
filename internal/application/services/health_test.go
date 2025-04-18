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
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/goforms/internal/application/services"
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
		wantBody    map[string]any
		wantLogCall bool
	}{
		{
			name:       "healthy service",
			pingError:  nil,
			wantStatus: http.StatusOK,
			wantBody: map[string]any{
				"success": true,
				"data": map[string]any{
					"status": "healthy",
				},
			},
			wantLogCall: false,
		},
		{
			name:       "unhealthy service",
			pingError:  errors.New("db connection failed"),
			wantStatus: http.StatusInternalServerError,
			wantBody: map[string]any{
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
			handler := services.NewHealthHandler(mockLogger, mockDB)

			// Create request and recorder
			req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
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

			var gotBody map[string]any
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
func deepEqual(t *testing.T, want, got map[string]any) bool {
	if len(want) != len(got) {
		return false
	}
	for key, wantVal := range want {
		gotVal, exists := got[key]
		if !exists {
			return false
		}
		switch v := wantVal.(type) {
		case map[string]any:
			if g, ok := gotVal.(map[string]any); !ok || !deepEqual(t, v, g) {
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
