package services_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/goforms/internal/application/services"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Mock types
type MockStore struct {
	expectations []func() error
}

func (m *MockStore) Get(ctx context.Context, id int64) (*subscription.Subscription, error) {
	if len(m.expectations) == 0 {
		return nil, nil
	}
	expect := m.expectations[0]
	m.expectations = m.expectations[1:]
	return nil, expect()
}

func (m *MockStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	if len(m.expectations) == 0 {
		return nil, nil
	}
	expect := m.expectations[0]
	m.expectations = m.expectations[1:]
	return nil, expect()
}

func (m *MockStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	if len(m.expectations) == 0 {
		return nil
	}
	expect := m.expectations[0]
	m.expectations = m.expectations[1:]
	return expect()
}

func (m *MockStore) Delete(ctx context.Context, id int64) error {
	if len(m.expectations) == 0 {
		return nil
	}
	expect := m.expectations[0]
	m.expectations = m.expectations[1:]
	return expect()
}

func (m *MockStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	if len(m.expectations) == 0 {
		return nil, nil
	}
	expect := m.expectations[0]
	m.expectations = m.expectations[1:]
	return nil, expect()
}

func (m *MockStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	if len(m.expectations) == 0 {
		return nil, nil
	}
	expect := m.expectations[0]
	m.expectations = m.expectations[1:]
	return nil, expect()
}

func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	if len(m.expectations) == 0 {
		return nil
	}
	expect := m.expectations[0]
	m.expectations = m.expectations[1:]
	return expect()
}

func (m *MockStore) ExpectGet(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockStore) ExpectGetByEmail(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockStore) ExpectCreate(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockStore) ExpectDelete(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockStore) ExpectGetByID(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockStore) ExpectList(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockStore) ExpectUpdateStatus(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockStore) Verify() error {
	if len(m.expectations) > 0 {
		return fmt.Errorf("unmet expectations: %d remaining", len(m.expectations))
	}
	return nil
}

type MockLogger struct {
	expectations []func() error
}

func NewMockLogger() logging.Logger {
	return &MockLogger{
		expectations: make([]func() error, 0),
	}
}

func (m *MockLogger) Info(msg string, fields ...logging.Field) {
	if len(m.expectations) > 0 {
		expect := m.expectations[0]
		m.expectations = m.expectations[1:]
		expect()
	}
}

func (m *MockLogger) Error(msg string, fields ...logging.Field) {
	if len(m.expectations) > 0 {
		expect := m.expectations[0]
		m.expectations = m.expectations[1:]
		expect()
	}
}

func (m *MockLogger) Debug(msg string, fields ...logging.Field) {
	if len(m.expectations) > 0 {
		expect := m.expectations[0]
		m.expectations = m.expectations[1:]
		expect()
	}
}

func (m *MockLogger) Warn(msg string, fields ...logging.Field) {
	if len(m.expectations) > 0 {
		expect := m.expectations[0]
		m.expectations = m.expectations[1:]
		expect()
	}
}

func (m *MockLogger) Int(key string, val int) logging.Field {
	return logging.Int(key, val)
}

func (m *MockLogger) Int32(key string, val int32) logging.Field {
	return logging.Any(key, val)
}

func (m *MockLogger) Int64(key string, val int64) logging.Field {
	return logging.Int64(key, val)
}

func (m *MockLogger) Uint(key string, val uint) logging.Field {
	return logging.Uint(key, val)
}

func (m *MockLogger) Uint32(key string, val uint32) logging.Field {
	return logging.Any(key, val)
}

func (m *MockLogger) Uint64(key string, val uint64) logging.Field {
	return logging.Any(key, val)
}

func (m *MockLogger) ExpectInfo(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockLogger) ExpectError(err error) {
	m.expectations = append(m.expectations, func() error { return err })
}

func (m *MockLogger) Verify() error {
	if len(m.expectations) > 0 {
		return fmt.Errorf("unmet expectations: %d remaining", len(m.expectations))
	}
	return nil
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
		setup          func(*MockStore, *MockLogger)
	}{
		{
			name:           "successful subscription",
			requestBody:    `{"email": "test@example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"status":"success","message":"Subscription created successfully"}`,
			setup: func(store *MockStore, logger *MockLogger) {
				store.ExpectGetByEmail(nil)
				store.ExpectCreate(nil)
				logger.ExpectInfo(nil)
			},
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "invalid-email"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"error","message":"Invalid email format"}`,
			setup: func(store *MockStore, logger *MockLogger) {
				logger.ExpectError(fmt.Errorf("invalid email format"))
			},
		},
		{
			name:           "duplicate email",
			requestBody:    `{"email": "existing@example.com"}`,
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"status":"error","message":"Email already subscribed"}`,
			setup: func(store *MockStore, logger *MockLogger) {
				store.ExpectGetByEmail(fmt.Errorf("email already subscribed"))
				logger.ExpectError(fmt.Errorf("email already subscribed"))
			},
		},
		{
			name:           "invalid request body",
			requestBody:    `{"invalid": "json"`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"error","message":"Invalid request body"}`,
			setup: func(store *MockStore, logger *MockLogger) {
				logger.ExpectError(fmt.Errorf("invalid request body"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockStore := &MockStore{}
			mockLogger := &MockLogger{}
			tt.setup(mockStore, mockLogger)

			// Create handler
			handler := NewSubscriptionHandler(mockStore, mockLogger)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/subscribe", strings.NewReader(tt.requestBody))
			rec := httptest.NewRecorder()

			// Handle request
			err := handler.HandleSubscribe(echo.New().NewContext(req, rec))
			require.NoError(t, err)

			// Verify response
			require.Equal(t, tt.expectedStatus, rec.Code)
			require.JSONEq(t, tt.expectedBody, rec.Body.String())

			// Verify mocks
			require.NoError(t, mockStore.Verify())
			require.NoError(t, mockLogger.Verify())
		})
	}
}
