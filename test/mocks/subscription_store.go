package mocks

import (
	"context"
	"errors"
	"fmt"

	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

// SubscriptionStore is a mock implementation of subscription.Store
type SubscriptionStore struct {
	mock.Mock
}

// NewSubscriptionStore creates a new mock subscription store
func NewSubscriptionStore() *SubscriptionStore {
	return &SubscriptionStore{}
}

// Create implements subscription.Store
func (m *SubscriptionStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// List implements subscription.Store
func (m *SubscriptionStore) List(ctx context.Context) ([]subscription.Subscription, error) {
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

// GetByID implements subscription.Store
func (m *SubscriptionStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	sub, ok := val.(*subscription.Subscription)
	if !ok {
		return nil, fmt.Errorf("unexpected type for subscription: %T", val)
	}
	return sub, args.Error(1)
}

// GetByEmail implements subscription.Store
func (m *SubscriptionStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	args := m.Called(ctx, email)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	sub, ok := val.(*subscription.Subscription)
	if !ok {
		return nil, fmt.Errorf("unexpected type for subscription: %T", val)
	}
	return sub, args.Error(1)
}

// UpdateStatus implements subscription.Store
func (m *SubscriptionStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// Delete implements subscription.Store
func (m *SubscriptionStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *SubscriptionStore) GetAll(ctx context.Context) ([]subscription.Subscription, error) {
	args := m.Called(ctx)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	subs, ok := val.([]subscription.Subscription)
	if !ok {
		return nil, fmt.Errorf("unexpected type for subscriptions: %T", val)
	}
	return subs, args.Error(1)
}
