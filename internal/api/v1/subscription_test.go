package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jonesrussell/goforms/internal/core/subscription"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		setupMock      func(*subscription.MockService)
	}{
		{
			name:           "valid subscription",
			payload:        `{"email":"test@example.com","name":"Test User"}`,
			expectedStatus: http.StatusCreated,
			setupMock: func(ms *subscription.MockService) {
				ms.On("CreateSubscription", mock.Anything, &subscription.Subscription{
					Email: "test@example.com",
					Name:  "Test User",
				}).Return(nil)
			},
		},
		{
			name:           "invalid email",
			payload:        `{"email":"invalid-email","name":"Test User"}`,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(ms *subscription.MockService) {},
		},
		{
			name:           "missing email",
			payload:        `{"name":"Test User"}`,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(ms *subscription.MockService) {},
		},
		{
			name:           "missing name",
			payload:        `{"email":"test@example.com"}`,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(ms *subscription.MockService) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(tt.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockService := new(subscription.MockService)
			tt.setupMock(mockService)

			mockLogger := logger.NewMockLogger()
			handler := NewSubscriptionAPI(mockService, mockLogger)

			// Test
			err := handler.CreateSubscription(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "data")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		setupMock      func(*subscription.MockService)
	}{
		{
			name:           "successful list",
			expectedStatus: http.StatusOK,
			setupMock: func(ms *subscription.MockService) {
				ms.On("ListSubscriptions", mock.Anything).Return([]subscription.Subscription{
					{ID: 1, Email: "test1@example.com"},
					{ID: 2, Email: "test2@example.com"},
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockService := new(subscription.MockService)
			tt.setupMock(mockService)

			mockLogger := logger.NewMockLogger()
			handler := NewSubscriptionAPI(mockService, mockLogger)

			// Test
			err := handler.ListSubscriptions(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "data")

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetSubscription(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expectedStatus int
		setupMock      func(*subscription.MockService)
	}{
		{
			name:           "existing subscription",
			id:             "1",
			expectedStatus: http.StatusOK,
			setupMock: func(ms *subscription.MockService) {
				ms.On("GetSubscription", mock.Anything, int64(1)).Return(&subscription.Subscription{
					ID:    1,
					Email: "test@example.com",
				}, nil)
			},
		},
		{
			name:           "non-existent subscription",
			id:             "999",
			expectedStatus: http.StatusNotFound,
			setupMock: func(ms *subscription.MockService) {
				ms.On("GetSubscription", mock.Anything, int64(999)).Return(nil, subscription.ErrSubscriptionNotFound)
			},
		},
		{
			name:           "invalid id",
			id:             "invalid",
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(ms *subscription.MockService) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
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
			err := handler.GetSubscription(c)

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
