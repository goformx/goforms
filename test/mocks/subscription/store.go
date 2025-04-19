package subscriptionmock

import (
	"context"
	"errors"

	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

// MockStore is a mock implementation of subscription.Store
type MockStore struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// List mocks the List method
func (m *MockStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	subs, ok := args.Get(0).([]subscription.Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion for subscriptions")
	}
	return subs, args.Error(1)
}

// GetByID mocks the GetByID method
func (m *MockStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sub, ok := args.Get(0).(*subscription.Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion for subscription")
	}
	return sub, args.Error(1)
}

// GetByEmail mocks the GetByEmail method
func (m *MockStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sub, ok := args.Get(0).(*subscription.Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion for subscription")
	}
	return sub, args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
