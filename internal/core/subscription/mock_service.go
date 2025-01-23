package subscription

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Service = (*MockService)(nil)

// MockService is a mock implementation of Service interface
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateSubscription(ctx context.Context, sub *Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockService) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Subscription), args.Error(1)
}

func (m *MockService) GetSubscription(ctx context.Context, id int64) (*Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Subscription), args.Error(1)
}

func (m *MockService) GetSubscriptionByEmail(ctx context.Context, email string) (*Subscription, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Subscription), args.Error(1)
}

func (m *MockService) UpdateSubscriptionStatus(ctx context.Context, id int64, status Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockService) DeleteSubscription(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
