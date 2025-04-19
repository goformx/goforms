package models

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/application/services"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

// Verify interface compliance at compile time
var _ subscription.Store = (*MockSubscriptionStore)(nil)
var _ services.PingContexter = (*MockPingContexter)(nil)

// MockSubscriptionStore is a mock implementation of subscription.Store
type MockSubscriptionStore struct {
	mock.Mock
}

// NewMockSubscriptionStore creates a new mock subscription store
func NewMockSubscriptionStore() *MockSubscriptionStore {
	return &MockSubscriptionStore{}
}

// Create implements subscription.Store
func (m *MockSubscriptionStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// List implements subscription.Store
func (m *MockSubscriptionStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	args := m.Called(ctx)
	subs, ok := args.Get(0).([]subscription.Subscription)
	if !ok {
		return nil, fmt.Errorf("invalid return type for List")
	}
	return subs, args.Error(1)
}

// Get implements subscription.Store
func (m *MockSubscriptionStore) Get(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	sub, ok := args.Get(0).(*subscription.Subscription)
	if !ok {
		return nil, fmt.Errorf("invalid return type for Get")
	}
	return sub, args.Error(1)
}

// GetByID implements subscription.Store
func (m *MockSubscriptionStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	return m.Get(ctx, id)
}

// GetByEmail implements subscription.Store
func (m *MockSubscriptionStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	args := m.Called(ctx, email)
	if sub := args.Get(0); sub != nil {
		return sub.(*subscription.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateStatus implements subscription.Store
func (m *MockSubscriptionStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// Delete implements subscription.Store
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
