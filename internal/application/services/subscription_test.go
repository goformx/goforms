package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/goforms/internal/application/services"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Mock types
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockStore) Get(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]subscription.Subscription), args.Error(1)
}

func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Int(key string, value int) logging.Field {
	args := m.Called(key, value)
	return args.Get(0).(logging.Field)
}

func (m *MockLogger) Int32(key string, value int32) logging.Field {
	args := m.Called(key, value)
	return args.Get(0).(logging.Field)
}

func (m *MockLogger) Int64(key string, value int64) logging.Field {
	args := m.Called(key, value)
	return args.Get(0).(logging.Field)
}

func (m *MockLogger) Uint(key string, value uint) logging.Field {
	args := m.Called(key, value)
	return args.Get(0).(logging.Field)
}

func (m *MockLogger) Uint32(key string, value uint32) logging.Field {
	args := m.Called(key, value)
	return args.Get(0).(logging.Field)
}

func (m *MockLogger) Uint64(key string, value uint64) logging.Field {
	args := m.Called(key, value)
	return args.Get(0).(logging.Field)
}

func TestSubscriptionHandler_HandleSubscribe(t *testing.T) {
	tests := []struct {
		name        string
		reqBody     string
		wantStatus  int
		wantBody    map[string]any
		wantLogCall bool
		setupMocks  func(*MockStore, *MockLogger, echo.Context)
	}{
		{
			name:       "success",
			reqBody:    `{"email":"test@example.com","name":"Test User"}`,
			wantStatus: http.StatusCreated,
			wantBody: map[string]any{
				"success": true,
				"data": map[string]any{
					"email": "test@example.com",
					"name":  "Test User",
				},
			},
			wantLogCall: true,
			setupMocks: func(store *MockStore, logger *MockLogger, c echo.Context) {
				store.On("GetByEmail", c.Request().Context(), "test@example.com").Return(nil, subscription.ErrSubscriptionNotFound)
				sub := &subscription.Subscription{
					Email:  "test@example.com",
					Name:   "Test User",
					Status: subscription.StatusPending,
				}
				store.On("Create", c.Request().Context(), sub).Return(nil)
				logger.On("Info", "subscription created", mock.Anything).Return()
			},
		},
		{
			name:       "invalid email",
			reqBody:    `{"email":"invalid"}`,
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]any{
				"success": false,
				"error":   "invalid email format",
			},
			wantLogCall: true,
			setupMocks: func(store *MockStore, logger *MockLogger, c echo.Context) {
				logger.On("Error", "invalid email format", mock.Anything).Return()
			},
		},
		{
			name:       "duplicate email",
			reqBody:    `{"email":"existing@example.com"}`,
			wantStatus: http.StatusConflict,
			wantBody: map[string]any{
				"success": false,
				"error":   "email already subscribed",
			},
			wantLogCall: true,
			setupMocks: func(store *MockStore, logger *MockLogger, c echo.Context) {
				store.On("GetByEmail", c.Request().Context(), "existing@example.com").Return(&subscription.Subscription{}, nil)
				logger.On("Error", "email already subscribed", mock.Anything).Return()
			},
		},
		{
			name:       "invalid request body",
			reqBody:    `invalid json`,
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]any{
				"success": false,
				"error":   "Invalid request body",
			},
			wantLogCall: true,
			setupMocks: func(store *MockStore, logger *MockLogger, c echo.Context) {
				logger.On("Error", "failed to bind subscription request", mock.Anything).Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			store := new(MockStore)
			logger := new(MockLogger)
			handler := services.NewSubscriptionHandler(store, logger)

			// Create request
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/subscribe", strings.NewReader(tt.reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.setupMocks != nil {
				tt.setupMocks(store, logger, c)
			}

			// Call handler
			err := handler.HandleSubscribe(c)
			if tt.wantStatus >= 400 {
				require.Error(t, err)
				he, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				assert.Equal(t, tt.wantStatus, he.Code)
				assert.Equal(t, tt.wantBody["error"], he.Message)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStatus, rec.Code)

				var gotBody map[string]any
				err = json.NewDecoder(rec.Body).Decode(&gotBody)
				require.NoError(t, err)
				assert.Equal(t, tt.wantBody, gotBody)
			}

			// Verify mock expectations
			store.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}
