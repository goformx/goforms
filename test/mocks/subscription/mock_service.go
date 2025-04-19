package subscriptionmock

import (
	"context"
	"errors"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/stretchr/testify/mock"
)

// Ensure SubscriptionService implements Service interface
var _ subscription.Service = (*SubscriptionService)(nil)

// SubscriptionService is a mock implementation of the Service interface
type SubscriptionService struct {
	mock.Mock
}

// NewSubscriptionService creates a new instance of SubscriptionService
func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{}
}

// CreateSubscription mocks the CreateSubscription method
func (m *SubscriptionService) CreateSubscription(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// ListSubscriptions mocks the ListSubscriptions method
func (m *SubscriptionService) ListSubscriptions(ctx context.Context) ([]subscription.Subscription, error) {
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

// GetSubscription mocks the GetSubscription method
func (m *SubscriptionService) GetSubscription(ctx context.Context, id int64) (*subscription.Subscription, error) {
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

// GetSubscriptionByEmail mocks the GetSubscriptionByEmail method
func (m *SubscriptionService) GetSubscriptionByEmail(
	ctx context.Context,
	email string,
) (*subscription.Subscription, error) {
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

// UpdateSubscriptionStatus mocks the UpdateSubscriptionStatus method
func (m *SubscriptionService) UpdateSubscriptionStatus(
	ctx context.Context,
	id int64,
	status subscription.Status,
) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// DeleteSubscription mocks the DeleteSubscription method
func (m *SubscriptionService) DeleteSubscription(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetByID implements Service.GetByID
func (m *SubscriptionService) GetByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, subscription.ErrSubscriptionNotFound
	}
	sub, ok := args.Get(0).(*subscription.Subscription)
	if !ok {
		return nil, subscription.ErrInvalidSubscription
	}
	return sub, args.Error(1)
}

// List implements Service.List
func (m *SubscriptionService) List(ctx context.Context) ([]*subscription.Subscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, nil
	}
	subs, ok := args.Get(0).([]*subscription.Subscription)
	if !ok {
		return nil, subscription.ErrInvalidSubscription
	}
	return subs, args.Error(1)
}
