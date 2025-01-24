package services_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/application/services"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	mocklog "github.com/jonesrussell/goforms/test/mocks/logging"
	storemock "github.com/jonesrussell/goforms/test/mocks/subscription"
)

func TestSubscriptionHandler_HandleSubscribe(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMocks     func(*storemock.MockStore, *mocklog.MockLogger)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful subscription",
			requestBody: `{
				"email": "test@example.com",
				"name": "Test User"
			}`,
			setupMocks: func(store *storemock.MockStore, logger *mocklog.MockLogger) {
				store.On("Create", mock.Anything, mock.MatchedBy(func(s *subscription.Subscription) bool {
					return s.Email == "test@example.com" && s.Name == "Test User"
				})).Return(nil)
				logger.On("Info", "subscription created", mock.Anything).Return()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"success":true,"data":{"id":0,"email":"test@example.com","name":"Test User","active":false,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:        "invalid request body",
			requestBody: `invalid json`,
			setupMocks: func(store *storemock.MockStore, logger *mocklog.MockLogger) {
				logger.On("Error", "failed to bind subscription request", mock.Anything).Return()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":"Invalid request body"}`,
		},
		{
			name: "missing required fields",
			requestBody: `{
				"email": "",
				"name": ""
			}`,
			setupMocks:     func(store *storemock.MockStore, logger *mocklog.MockLogger) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":"email is required"}`,
		},
		{
			name: "store error",
			requestBody: `{
				"email": "test@example.com",
				"name": "Test User"
			}`,
			setupMocks: func(store *storemock.MockStore, logger *mocklog.MockLogger) {
				store.On("Create", mock.Anything, mock.Anything).Return(errors.New("store error"))
				logger.On("Error", "failed to create subscription", mock.Anything).Return()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"success":false,"error":"Failed to create subscription"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/subscribe", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			store := new(storemock.MockStore)
			logger := new(mocklog.MockLogger)
			tt.setupMocks(store, logger)

			handler := services.NewSubscriptionHandler(store, logger)

			// Test
			err := handler.HandleSubscribe(c)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			store.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}
