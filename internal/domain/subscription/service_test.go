package subscription_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

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
		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, nil)
		mockStore.ExpectCreate(context.Background(), &subscription.Subscription{
			Name:  "Test User",
			Email: "test@example.com",
		}, nil)

		service := subscription.NewService(mockLogger, mockStore)

		sub := &subscription.Subscription{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := service.CreateSubscription(context.Background(), sub)
		assert.NoError(t, err)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("duplicate_email", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", &subscription.Subscription{}, nil)

		service := subscription.NewService(mockLogger, mockStore)

		sub := &subscription.Subscription{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := service.CreateSubscription(context.Background(), sub)
		assert.Error(t, err)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}

func TestListSubscriptions(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	expected := []subscription.Subscription{
		{ID: 1, Name: "Test User 1", Email: "test1@example.com"},
		{ID: 2, Name: "Test User 2", Email: "test2@example.com"},
	}

	mockStore := subscriptionmock.NewMockStore()
	mockStore.ExpectList(context.Background(), expected, nil)

	service := subscription.NewService(mockLogger, mockStore)

	subs, err := service.ListSubscriptions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, subs)
	if err := mockStore.Verify(); err != nil {
		t.Errorf("store expectations not met: %v", err)
	}
	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}

func TestGetSubscription(t *testing.T) {
	t.Run("existing_subscription", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		expected := &subscription.Subscription{
			ID:    123,
			Name:  "Test User",
			Email: "test@example.com",
		}

		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByID(context.Background(), int64(123), expected, nil)

		service := subscription.NewService(mockLogger, mockStore)

		sub, err := service.GetSubscription(context.Background(), 123)
		assert.NoError(t, err)
		assert.Equal(t, expected, sub)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("non-existent_subscription", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockLogger.ExpectError("failed to get subscription",
			logging.Error(subscription.ErrSubscriptionNotFound),
		)

		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByID(context.Background(), int64(123), nil, subscription.ErrSubscriptionNotFound)

		service := subscription.NewService(mockLogger, mockStore)

		sub, err := service.GetSubscription(context.Background(), 123)
		assert.Error(t, err)
		assert.Nil(t, sub)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}

func TestUpdateSubscriptionStatus(t *testing.T) {
	mockStore := subscriptionmock.NewMockStore()
	mockLogger := mocklogging.NewMockLogger()
	service := subscription.NewService(mockLogger, mockStore)

	tests := []struct {
		name    string
		id      int64
		status  subscription.Status
		setup   func(*subscriptionmock.MockStore)
		wantErr error
	}{
		{
			name:   "valid status update",
			id:     1,
			status: subscription.StatusActive,
			setup: func(store *subscriptionmock.MockStore) {
				store.ExpectGetByID(context.Background(), int64(1), &subscription.Subscription{ID: 1}, nil)
				store.ExpectUpdateStatus(context.Background(), int64(1), subscription.StatusActive, nil)
			},
			wantErr: nil,
		},
		{
			name:   "invalid status",
			id:     1,
			status: "invalid",
			setup: func(store *subscriptionmock.MockStore) {
			},
			wantErr: subscription.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockStore.Reset()
			mockLogger.Reset()

			// Setup mock expectations
			tt.setup(mockStore)

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
			if err := mockStore.Verify(); err != nil {
				t.Errorf("store expectations not met: %v", err)
			}
			if err := mockLogger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	t.Run("successful_deletion", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByID(context.Background(), int64(1), &subscription.Subscription{ID: 1}, nil)
		mockStore.ExpectDelete(context.Background(), int64(1), nil)

		service := subscription.NewService(mockLogger, mockStore)

		err := service.DeleteSubscription(context.Background(), 1)
		assert.NoError(t, err)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("non-existent_subscription", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockLogger.ExpectError("failed to get subscription",
			logging.Error(subscription.ErrSubscriptionNotFound),
		)

		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByID(context.Background(), int64(1), nil, subscription.ErrSubscriptionNotFound)

		service := subscription.NewService(mockLogger, mockStore)

		err := service.DeleteSubscription(context.Background(), 1)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, subscription.ErrSubscriptionNotFound))
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}

func TestGetSubscriptionByEmail(t *testing.T) {
	t.Run("existing_subscription", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		expected := &subscription.Subscription{
			ID:     1,
			Email:  "test@example.com",
			Name:   "Test User",
			Status: subscription.StatusActive,
		}

		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", expected, nil)

		service := subscription.NewService(mockLogger, mockStore)

		sub, err := service.GetSubscriptionByEmail(context.Background(), "test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, expected, sub)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("non-existent_subscription", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockLogger.ExpectError("failed to get subscription by email",
			logging.Error(subscription.ErrSubscriptionNotFound),
		)

		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, subscription.ErrSubscriptionNotFound)

		service := subscription.NewService(mockLogger, mockStore)

		sub, err := service.GetSubscriptionByEmail(context.Background(), "test@example.com")
		assert.Error(t, err)
		assert.Nil(t, sub)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("store_error", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockLogger.ExpectError("failed to get subscription by email",
			logging.Error(errors.New("database error")),
		)

		mockStore := subscriptionmock.NewMockStore()
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, errors.New("database error"))

		service := subscription.NewService(mockLogger, mockStore)

		sub, err := service.GetSubscriptionByEmail(context.Background(), "test@example.com")
		assert.Error(t, err)
		assert.Equal(t, "failed to get subscription by email: database error", err.Error())
		assert.Nil(t, sub)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("empty_email", func(t *testing.T) {
		mockLogger := mocklogging.NewMockLogger()
		mockStore := subscriptionmock.NewMockStore()

		service := subscription.NewService(mockLogger, mockStore)

		sub, err := service.GetSubscriptionByEmail(context.Background(), "")
		assert.Error(t, err)
		assert.Equal(t, "invalid input: email is required", err.Error())
		assert.Nil(t, sub)
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}
