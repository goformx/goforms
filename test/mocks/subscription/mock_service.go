package subscriptionmock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

// Ensure MockService implements Service interface
var _ subscription.Service = (*MockService)(nil)

// MockService is a mock implementation of the Service interface
type MockService struct {
	mock.Mock
}

// NewMockService creates a new instance of MockService
func NewMockService() *MockService {
	return &MockService{}
}

// CreateSubscription mocks the CreateSubscription method
func (m *MockService) CreateSubscription(ctx context.Context, sub *subscription.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// ListSubscriptions mocks the ListSubscriptions method
func (m *MockService) ListSubscriptions(ctx context.Context) ([]subscription.Subscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]subscription.Subscription), args.Error(1)
}

// GetSubscription mocks the GetSubscription method
func (m *MockService) GetSubscription(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

// GetSubscriptionByEmail mocks the GetSubscriptionByEmail method
func (m *MockService) GetSubscriptionByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

// UpdateSubscriptionStatus mocks the UpdateSubscriptionStatus method
func (m *MockService) UpdateSubscriptionStatus(ctx context.Context, id int64, status subscription.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// DeleteSubscription mocks the DeleteSubscription method
func (m *MockService) DeleteSubscription(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
