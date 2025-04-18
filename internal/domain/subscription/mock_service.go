package subscription

import (
	"context"
	"errors"

	"github.com/stretchr/testify/mock"
)

// Verify interface compliance at compile time
var _ Service = (*MockService)(nil)

// MockService is a mock implementation of Service interface
type MockService struct {
	mock.Mock
}

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{}
}

// CreateSubscription mocks the CreateSubscription method
func (m *MockService) CreateSubscription(ctx context.Context, sub *Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// ListSubscriptions mocks the ListSubscriptions method
func (m *MockService) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	subs, ok := args.Get(0).([]Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion")
	}
	return subs, nil
}

// GetSubscription mocks the GetSubscription method
func (m *MockService) GetSubscription(ctx context.Context, id int64) (*Subscription, error) {
	args := m.Called(ctx, id)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	sub, ok := args.Get(0).(*Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion")
	}
	return sub, nil
}

// GetSubscriptionByEmail mocks the GetSubscriptionByEmail method
func (m *MockService) GetSubscriptionByEmail(ctx context.Context, email string) (*Subscription, error) {
	args := m.Called(ctx, email)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	sub, ok := args.Get(0).(*Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion")
	}
	return sub, nil
}

// UpdateSubscriptionStatus mocks the UpdateSubscriptionStatus method
func (m *MockService) UpdateSubscriptionStatus(ctx context.Context, id int64, status Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// DeleteSubscription mocks the DeleteSubscription method
func (m *MockService) DeleteSubscription(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) GetSubscriptionByID(ctx context.Context, id uint) (*Subscription, error) {
	args := m.Called(ctx, id)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	sub, ok := args.Get(0).(*Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion")
	}
	return sub, nil
}
