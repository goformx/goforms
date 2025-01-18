package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/core/subscription"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/response"
)

func TestSubscriptionIntegration(t *testing.T) {
	// Setup
	e := echo.New()
	mockStore := subscription.NewMockStore()
	mockLogger := logger.NewMockLogger()

	handler := NewSubscriptionHandler(mockStore, mockLogger)
	handler.Register(e)

	tests := []struct {
		name           string
		subscription   subscription.Subscription
		setupMock      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "valid subscription",
			subscription: subscription.Subscription{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setupMock: func() {
				mockStore.On("Create", mock.Anything, &subscription.Subscription{
					Email: "test@example.com",
					Name:  "Test User",
				}).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
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

			req := httptest.NewRequest(http.MethodPost, "/api/v1/subscribe", bytes.NewReader(jsonData))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			// Verify response code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse response
			var resp response.Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tc.expectedError {
				assert.Equal(t, "error", resp.Status)
				assert.NotEmpty(t, resp.Message)
			} else {
				assert.Equal(t, "success", resp.Status)

				// Verify subscription data
				subscriptionData, err := json.Marshal(resp.Data)
				assert.NoError(t, err)
				var subscription subscription.Subscription
				err = json.Unmarshal(subscriptionData, &subscription)
				assert.NoError(t, err)
				assert.Equal(t, tc.subscription.Email, subscription.Email)
				assert.Equal(t, tc.subscription.Name, subscription.Name)
			}

			// Verify logger and mock expectations
			if !tc.expectedError {
				assert.True(t, mockLogger.HasInfoLog("subscription created"))
			}
			mockStore.AssertExpectations(t)
		})
	}
}
