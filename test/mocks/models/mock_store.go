package models

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/application/services"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

// Verify interface compliance at compile time
var _ subscription.Store = (*MockSubscriptionStore)(nil)
var _ services.PingContexter = (*MockPingContexter)(nil)

// MockSubscriptionStore is a mock implementation of subscription.Store interface
type MockSubscriptionStore struct {
	mock.Mock
}

// NewMockSubscriptionStore creates a new mock subscription store
func NewMockSubscriptionStore() *MockSubscriptionStore {
	return &MockSubscriptionStore{}
}

// Create mocks the Create method of the subscription.Store interface.
// It records the subscription and returns the configured error.
func (m *MockSubscriptionStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// List mocks the List method of the subscription.Store interface.
// It returns the configured list of subscriptions and error.
func (m *MockSubscriptionStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]subscription.Subscription), args.Error(1)
}

// GetByID mocks the GetByID method of the subscription.Store interface.
// It returns the configured subscription and error based on the provided ID.
func (m *MockSubscriptionStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

// GetByEmail mocks the GetByEmail method of the subscription.Store interface.
// It returns the configured subscription and error based on the provided email.
func (m *MockSubscriptionStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method of the subscription.Store interface.
// It records the status update and returns the configured error.
func (m *MockSubscriptionStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// Delete mocks the Delete method of the subscription.Store interface.
// It records the deletion and returns the configured error.
func (m *MockSubscriptionStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPingContexter is a mock implementation of services.PingContexter interface
type MockPingContexter struct {
	mock.Mock
}

// NewMockPingContexter creates a new mock ping contexter
func NewMockPingContexter() *MockPingContexter {
	return &MockPingContexter{}
}

// PingContext mocks the PingContext method of the services.PingContexter interface.
// It records the context and returns the configured error.
func (m *MockPingContexter) PingContext(c echo.Context) error {
	args := m.Called(c)
	return args.Error(0)
}
