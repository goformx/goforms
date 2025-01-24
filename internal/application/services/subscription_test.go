package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/core/subscription"
	"github.com/jonesrussell/goforms/internal/logger"
	storemock "github.com/jonesrussell/goforms/test/mocks/store/subscription"
)

func TestSubscriptionHandler_HandleSubscribe(t *testing.T) {
	e := echo.New()
	mockLogger := logger.NewMockLogger()
	mockStore := storemock.NewMockStore()
	handler := NewSubscriptionHandler(mockStore, mockLogger)

	tests := []struct {
		name           string
		subscription   *subscription.Subscription
		setupMock      func(*storemock.MockStore)
		expectedStatus int
	}{
		{
			name: "valid subscription",
			subscription: &subscription.Subscription{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setupMock: func(mockStore *storemock.MockStore) {
				mockStore.On("Create", mock.Anything, &subscription.Subscription{
					Email: "test@example.com",
					Name:  "Test User",
				}).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockStore)

			// Create request
			jsonData, err := json.Marshal(tt.subscription)
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
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Verify mock expectations
			mockStore.AssertExpectations(t)
		})
	}
}
