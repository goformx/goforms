package services_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/application/services"
	mocklog "github.com/jonesrussell/goforms/test/mocks/logging"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) PingContext(ctx echo.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHealthHandler_HandleHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		dbError        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "healthy service",
			dbError:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"healthy"}`,
		},
		{
			name:           "unhealthy service",
			dbError:        errors.New("connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"success":false,"error":"Service is not healthy"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockDB := new(mockDB)
			mockDB.On("PingContext", c).Return(tt.dbError)

			logger := new(mocklog.MockLogger)
			if tt.dbError != nil {
				logger.On("Error", "health check failed", mock.Anything).Return()
			}

			handler := services.NewHealthHandler(logger, mockDB)

			// Test
			err := handler.HandleHealthCheck(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			mockDB.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}
