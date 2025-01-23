package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jonesrussell/goforms/internal/core/subscription"
	"github.com/jonesrussell/goforms/internal/logger"
	subscriptionmock "github.com/jonesrussell/goforms/test/mocks/subscription"
	"github.com/jonesrussell/goforms/test/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		name           string
		subscription   subscription.Subscription
		setupFn        func(*subscriptionmock.MockService)
		expectedStatus int
	}{
		{
			name: "valid subscription",
			subscription: subscription.Subscription{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setupFn: func(ms *subscriptionmock.MockService) {
				ms.On("CreateSubscription", mock.Anything, mock.MatchedBy(func(s *subscription.Subscription) bool {
					return s.Email == "test@example.com" && s.Name == "Test User"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid email",
			subscription: subscription.Subscription{
				Email: "invalid-email",
				Name:  "Test User",
			},
			setupFn:        func(ms *subscriptionmock.MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			subscription: subscription.Subscription{
				Name: "Test User",
			},
			setupFn:        func(ms *subscriptionmock.MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing name",
			subscription: subscription.Subscription{
				Email: "test@example.com",
			},
			setupFn:        func(ms *subscriptionmock.MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setup := utils.NewTestSetup()
			defer setup.Close()

			mockService := subscriptionmock.NewMockService()
			tt.setupFn(mockService)

			api := NewSubscriptionAPI(mockService, setup.Logger)

			// Create request
			req, err := utils.NewJSONRequest(http.MethodPost, "/api/v1/subscriptions", tt.subscription)
			assert.NoError(t, err)

			// Execute request
			c, rec := utils.NewTestContext(setup.Echo, req)
			err = api.CreateSubscription(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusCreated {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	tests := []struct {
		name           string
		setupFn        func(*subscriptionmock.MockService)
		expectedStatus int
	}{
		{
			name: "successful list",
			setupFn: func(ms *subscriptionmock.MockService) {
				ms.On("ListSubscriptions", mock.Anything).Return([]subscription.Subscription{
					{ID: 1, Email: "test1@example.com"},
					{ID: 2, Email: "test2@example.com"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupFn: func(ms *subscriptionmock.MockService) {
				ms.On("ListSubscriptions", mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setup := utils.NewTestSetup()
			defer setup.Close()

			mockService := subscriptionmock.NewMockService()
			tt.setupFn(mockService)

			api := NewSubscriptionAPI(mockService, setup.Logger)

			// Create request
			req, err := utils.NewJSONRequest(http.MethodGet, "/api/v1/subscriptions", nil)
			assert.NoError(t, err)

			// Execute request
			c, rec := utils.NewTestContext(setup.Echo, req)
			err = api.ListSubscriptions(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusOK {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetSubscription(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupFn        func(*subscriptionmock.MockService)
		expectedStatus int
	}{
		{
			name: "existing subscription",
			id:   "1",
			setupFn: func(ms *subscriptionmock.MockService) {
				ms.On("GetSubscription", mock.Anything, int64(1)).Return(&subscription.Subscription{
					ID:    1,
					Email: "test@example.com",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "non-existent subscription",
			id:   "999",
			setupFn: func(ms *subscriptionmock.MockService) {
				ms.On("GetSubscription", mock.Anything, int64(999)).Return(nil, subscription.ErrSubscriptionNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			id:             "invalid",
			setupFn:        func(ms *subscriptionmock.MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setup := utils.NewTestSetup()
			defer setup.Close()

			mockService := subscriptionmock.NewMockService()
			tt.setupFn(mockService)

			api := NewSubscriptionAPI(mockService, setup.Logger)

			// Create request
			req, err := utils.NewJSONRequest(http.MethodGet, "/", nil)
			assert.NoError(t, err)

			// Execute request
			c, rec := utils.NewTestContext(setup.Echo, req)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err = api.GetSubscription(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusOK {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateSubscriptionStatus(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		payload        string
		expectedStatus int
		setupMock      func(*subscription.MockService)
	}{
		{
			name:           "valid update",
			id:             "1",
			payload:        `{"status":"active"}`,
			expectedStatus: http.StatusOK,
			setupMock: func(ms *subscription.MockService) {
				ms.On("UpdateSubscriptionStatus", mock.Anything, int64(1), subscription.StatusActive).Return(nil)
			},
		},
		{
			name:           "invalid status",
			id:             "1",
			payload:        `{"status":"invalid"}`,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(ms *subscription.MockService) {},
		},
		{
			name:           "non-existent subscription",
			id:             "999",
			payload:        `{"status":"active"}`,
			expectedStatus: http.StatusNotFound,
			setupMock: func(ms *subscription.MockService) {
				ms.On("UpdateSubscriptionStatus", mock.Anything, int64(999), subscription.StatusActive).Return(subscription.ErrSubscriptionNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(tt.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/subscriptions/:id/status")
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			mockService := new(subscription.MockService)
			tt.setupMock(mockService)

			mockLogger := logger.NewMockLogger()
			handler := NewSubscriptionAPI(mockService, mockLogger)

			// Test
			err := handler.UpdateSubscriptionStatus(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "data")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expectedStatus int
		setupMock      func(*subscription.MockService)
	}{
		{
			name:           "successful delete",
			id:             "1",
			expectedStatus: http.StatusOK,
			setupMock: func(ms *subscription.MockService) {
				ms.On("DeleteSubscription", mock.Anything, int64(1)).Return(nil)
			},
		},
		{
			name:           "non-existent subscription",
			id:             "999",
			expectedStatus: http.StatusNotFound,
			setupMock: func(ms *subscription.MockService) {
				ms.On("DeleteSubscription", mock.Anything, int64(999)).Return(subscription.ErrSubscriptionNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/subscriptions/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			mockService := new(subscription.MockService)
			tt.setupMock(mockService)

			mockLogger := logger.NewMockLogger()
			handler := NewSubscriptionAPI(mockService, mockLogger)

			// Test
			err := handler.DeleteSubscription(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			mockService.AssertExpectations(t)
		})
	}
}

func TestRegister(t *testing.T) {
	// Setup
	setup := utils.NewTestSetup()
	defer setup.Close()

	mockService := subscriptionmock.NewMockService()

	// Set up mock expectations for any potential service calls
	mockService.On("CreateSubscription", mock.Anything, mock.Anything).Return(nil)
	mockService.On("ListSubscriptions", mock.Anything).Return([]subscription.Subscription{}, nil)
	mockService.On("GetSubscription", mock.Anything, mock.Anything).Return(&subscription.Subscription{}, nil)

	api := NewSubscriptionAPI(mockService, setup.Logger)

	// Test registration
	api.Register(setup.Echo)

	// Verify routes are registered by making test requests
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/subscriptions"},
		{http.MethodGet, "/api/v1/subscriptions"},
		{http.MethodGet, "/api/v1/subscriptions/1"},
	}

	for _, route := range routes {
		req, err := utils.NewJSONRequest(route.method, route.path, nil)
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		setup.Echo.ServeHTTP(rec, req)
		assert.NotEqual(t, http.StatusNotFound, rec.Code, "Route %s %s should exist", route.method, route.path)
	}
}
