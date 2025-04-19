package services_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/goforms/internal/application/services"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

var (
	ErrNotFound = errors.New("subscription not found")
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Get(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, ErrNotFound
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, ErrNotFound
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, ErrNotFound
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
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

func NewSubscriptionHandler(store subscription.Store, logger logging.Logger) *services.SubscriptionHandler {
	return services.NewSubscriptionHandler(store, logger)
}

func TestSubscriptionHandler_HandleSubscribe(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedBody   string
		setup          func(*MockStore, *mocklogging.MockLogger)
	}{
		{
			name:           "successful subscription",
			requestBody:    `{"email": "test@example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"status":"success","message":"Subscription created successfully"}`,
			setup: func(store *MockStore, logger *mocklogging.MockLogger) {
				store.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				store.On("Create", mock.Anything, mock.Anything).Return(nil)
				logger.ExpectInfo("subscription created")
			},
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "invalid-email"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"error","message":"Invalid email format"}`,
			setup: func(store *MockStore, logger *mocklogging.MockLogger) {
				logger.ExpectError("invalid email format")
			},
		},
		{
			name:           "duplicate email",
			requestBody:    `{"email": "existing@example.com"}`,
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"status":"error","message":"Email already subscribed"}`,
			setup: func(store *MockStore, logger *mocklogging.MockLogger) {
				store.On("GetByEmail", mock.Anything, "existing@example.com").Return(&subscription.Subscription{}, nil)
				logger.ExpectError("email already subscribed")
			},
		},
		{
			name:           "empty email",
			requestBody:    `{"email": ""}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"error","message":"Email is required"}`,
			setup: func(store *MockStore, logger *mocklogging.MockLogger) {
				logger.ExpectError("email is required")
			},
		},
		{
			name:           "missing email",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"error","message":"Email is required"}`,
			setup: func(store *MockStore, logger *mocklogging.MockLogger) {
				logger.ExpectError("email is required")
			},
		},
		{
			name:           "invalid json",
			requestBody:    `{`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"error","message":"Invalid request body"}`,
			setup: func(store *MockStore, logger *mocklogging.MockLogger) {
				logger.ExpectError("invalid request body")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockStore := &MockStore{}
			mockLogger := mocklogging.NewMockLogger()
			tt.setup(mockStore, mockLogger)

			handler := NewSubscriptionHandler(mockStore, mockLogger)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			// Execute
			err := handler.HandleSubscribe(c)

			// Assert
			require.NoError(t, err)
			require.Equal(t, tt.expectedStatus, rec.Code)
			require.JSONEq(t, tt.expectedBody, rec.Body.String())
			mockStore.AssertExpectations(t)
			require.NoError(t, mockLogger.Verify())
		})
	}
}
