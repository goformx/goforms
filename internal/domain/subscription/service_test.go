package subscription_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	subscriptionmock "github.com/jonesrussell/goforms/test/mocks/store/subscription"
)

func TestNewService(t *testing.T) {
	mockStore := subscriptionmock.NewMockStore()
	mockLogger := mocklogging.NewMockLogger()
	service := subscription.NewService(mockLogger, mockStore)
	assert.NotNil(t, service)
}

func TestCreateSubscription(t *testing.T) {
	t.Run("valid_subscription", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockLogger.ExpectInfo("creating subscription",
			logging.String("email", "test@example.com"),
			logging.String("name", "Test User"),
		)
		mockLogger.ExpectInfo("subscription created",
			logging.String("email", "test@example.com"),
		)

		mockStore := subscriptionmock.NewMockStore()
		mockStore.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
		mockStore.On("Create", mock.Anything, mock.Anything).Return(nil)

		service := subscription.NewService(mockLogger, mockStore)

		sub := &subscription.Subscription{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := service.CreateSubscription(context.Background(), sub)
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockLogger.ExpectInfo("creating subscription",
			logging.String("email", "test@example.com"),
			logging.String("name", "Test User"),
		)
		mockLogger.ExpectError("failed to create subscription",
			logging.Error(errors.New("email already exists")),
		)

		mockStore := subscriptionmock.NewMockStore()
		mockStore.On("GetByEmail", mock.Anything, "test@example.com").Return(&subscription.Subscription{}, nil)

		service := subscription.NewService(mockLogger, mockStore)

		sub := &subscription.Subscription{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := service.CreateSubscription(context.Background(), sub)
		assert.Error(t, err)
		mockStore.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestListSubscriptions(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("listing subscriptions")
	mockLogger.ExpectInfo("subscriptions retrieved",
		logging.Int("count", 2),
	)

	expected := []subscription.Subscription{
		{ID: 1, Name: "Test User 1", Email: "test1@example.com"},
		{ID: 2, Name: "Test User 2", Email: "test2@example.com"},
	}

	mockStore := subscriptionmock.NewMockStore()
	mockStore.On("List", mock.Anything).Return(expected, nil)

	service := subscription.NewService(mockLogger, mockStore)

	subs, err := service.ListSubscriptions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, subs)
	mockStore.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestGetSubscription(t *testing.T) {
	t.Run("existing_subscription", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockLogger.ExpectInfo("getting subscription",
			logging.Int("id", 123),
		)
		mockLogger.ExpectInfo("subscription retrieved",
			logging.Int("id", 123),
		)

		expected := &subscription.Subscription{
			ID:    123,
			Name:  "Test User",
			Email: "test@example.com",
		}

		mockStore := subscriptionmock.NewMockStore()
		mockStore.On("GetByID", mock.Anything, int64(123)).Return(expected, nil)

		service := subscription.NewService(mockLogger, mockStore)

		sub, err := service.GetSubscription(context.Background(), 123)
		assert.NoError(t, err)
		assert.Equal(t, expected, sub)
		mockStore.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("non-existent_subscription", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockLogger.ExpectInfo("getting subscription",
			logging.Int("id", 123),
		)
		mockLogger.ExpectError("failed to get subscription",
			logging.Error(errors.New("subscription not found")),
			logging.Int("id", 123),
		)

		mockStore := subscriptionmock.NewMockStore()
		mockStore.On("GetByID", mock.Anything, int64(123)).Return(nil, errors.New("subscription not found"))

		service := subscription.NewService(mockLogger, mockStore)

		sub, err := service.GetSubscription(context.Background(), 123)
		assert.Error(t, err)
		assert.Nil(t, sub)
		mockStore.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestUpdateSubscriptionStatus(t *testing.T) {
	// Create mock store
	mockStore := subscriptionmock.NewMockStore()
	mockLogger := mocklogging.NewMockLogger()

	// Create service with mocks
	service := subscription.NewService(mockLogger, mockStore)

	// Test cases
	tests := []struct {
		name    string
		id      int64
		status  subscription.Status
		setup   func()
		wantErr error
	}{
		{
			name:   "valid status update",
			id:     1,
			status: subscription.StatusActive,
			setup: func() {
				mockLogger.ExpectInfo("updating subscription status",
					logging.Int("id", 1),
					logging.String("status", string(subscription.StatusActive)),
				)
				mockLogger.ExpectInfo("subscription status updated",
					logging.Int("id", 1),
					logging.String("status", string(subscription.StatusActive)),
				)
				mockStore.On("GetByID", mock.Anything, int64(1)).
					Return(&subscription.Subscription{ID: 1}, nil)
				mockStore.On("UpdateStatus", mock.Anything, int64(1), subscription.StatusActive).
					Return(nil)
			},
		},
		{
			name:   "non-existent subscription",
			id:     999,
			status: subscription.StatusActive,
			setup: func() {
				mockLogger.ExpectInfo("updating subscription status",
					logging.Int("id", 999),
					logging.String("status", string(subscription.StatusActive)),
				)
				mockLogger.ExpectError("failed to update subscription status",
					logging.Error(subscription.ErrSubscriptionNotFound),
					logging.Int("id", 999),
				)
				mockStore.On("GetByID", mock.Anything, int64(999)).
					Return(nil, subscription.ErrSubscriptionNotFound)
			},
			wantErr: subscription.ErrSubscriptionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.setup()

			// Call service method
			err := service.UpdateSubscriptionStatus(context.Background(), tt.id, tt.status)

			// Assert error
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockStore.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	// Create mock store
	mockStore := subscriptionmock.NewMockStore()
	mockLogger := mocklogging.NewMockLogger()

	// Create service with mocks
	service := subscription.NewService(mockLogger, mockStore)

	// Test cases
	tests := []struct {
		name    string
		id      int64
		setup   func()
		wantErr error
	}{
		{
			name: "existing subscription",
			id:   1,
			setup: func() {
				mockLogger.ExpectInfo("deleting subscription",
					logging.Int("id", 1),
				)
				mockLogger.ExpectInfo("subscription deleted",
					logging.Int("id", 1),
				)
				mockStore.On("GetByID", mock.Anything, int64(1)).
					Return(&subscription.Subscription{ID: 1}, nil)
				mockStore.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
		},
		{
			name: "non-existent subscription",
			id:   999,
			setup: func() {
				mockLogger.ExpectInfo("deleting subscription",
					logging.Int("id", 999),
				)
				mockLogger.ExpectError("failed to delete subscription",
					logging.Error(subscription.ErrSubscriptionNotFound),
					logging.Int("id", 999),
				)
				mockStore.On("GetByID", mock.Anything, int64(999)).
					Return(nil, subscription.ErrSubscriptionNotFound)
			},
			wantErr: subscription.ErrSubscriptionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.setup()

			// Call service method
			err := service.DeleteSubscription(context.Background(), tt.id)

			// Assert error
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockStore.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetSubscriptionByEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		setup   func(*subscriptionmock.MockStore)
		want    *subscription.Subscription
		wantErr error
	}{
		{
			name:  "existing subscription",
			email: "test@example.com",
			setup: func(ms *subscriptionmock.MockStore) {
				sub := &subscription.Subscription{
					ID:     1,
					Name:   "Test User",
					Email:  "test@example.com",
					Status: subscription.StatusActive,
				}
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(sub, nil)
			},
			want: &subscription.Subscription{
				ID:     1,
				Name:   "Test User",
				Email:  "test@example.com",
				Status: subscription.StatusActive,
			},
			wantErr: nil,
		},
		{
			name:  "non-existent subscription",
			email: "nonexistent@example.com",
			setup: func(ms *subscriptionmock.MockStore) {
				ms.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, subscription.ErrSubscriptionNotFound)
			},
			want:    nil,
			wantErr: subscription.ErrSubscriptionNotFound,
		},
		{
			name:  "store error",
			email: "test@example.com",
			setup: func(ms *subscriptionmock.MockStore) {
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("database error"))
			},
			want:    nil,
			wantErr: errors.New("database error"),
		},
		{
			name:  "empty email",
			email: "",
			setup: func(ms *subscriptionmock.MockStore) {
				// No mock setup needed as it should fail before store call
			},
			want:    nil,
			wantErr: subscription.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock store for each test
			mockStore := subscriptionmock.NewMockStore()
			mockLogger := mocklogging.NewMockLogger()
			mockLogger.ExpectInfo("getting subscription by email")

			// Setup mock expectations
			tt.setup(mockStore)

			// Create service with mocks
			service := subscription.NewService(mockLogger, mockStore)

			// Call method
			got, err := service.GetSubscriptionByEmail(context.Background(), tt.email)

			// Assert error
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Assert result
			assert.Equal(t, tt.want, got)

			// Verify mock expectations
			mockStore.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
