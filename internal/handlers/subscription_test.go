package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionHandler_HandleSubscribe(t *testing.T) {
	e := echo.New()
	mockLogger := logger.NewMockLogger()
	mockStore := models.NewMockSubscriptionStore()
	handler := NewSubscriptionHandler(mockStore, mockLogger)

	tests := []struct {
		name           string
		subscription   *models.Subscription
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "valid subscription",
			subscription: &models.Subscription{
				Email: "test@example.com",
			},
			setupMock: func() {
				mockStore.On("Create", &models.Subscription{Email: "test@example.com"}).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock between tests
			mockStore.ExpectedCalls = nil
			mockStore.Calls = nil

			// Setup mock expectations
			tc.setupMock()

			// Create request
			jsonData, err := json.Marshal(tc.subscription)
			assert.NoError(t, err)
			body := string(jsonData)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/subscribe", strings.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call handler
			err = handler.HandleSubscribe(c)
			assert.NoError(t, err)

			// Assert response
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Verify mock expectations
			mockStore.AssertExpectations(t)
		})
	}
}
