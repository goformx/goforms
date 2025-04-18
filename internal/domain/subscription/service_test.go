package subscription_test

import (
	"errors"
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	subscriptionmock "github.com/jonesrussell/goforms/test/mocks/store/subscription"
)

func TestNewService(t *testing.T) {
	mockStore := subscriptionmock.NewMockStore(t)
	mockLogger := mocklogging.NewMockLogger()

	service := subscription.NewService(mockStore, mockLogger)
	if service == nil {
		t.Error("expected service to be created")
	}
}

func TestCreateSubscription(t *testing.T) {
	t.Run("valid_subscription", func(t *testing.T) {
		mockStore := subscriptionmock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()

		mockStore.ExpectGetByEmail(t.Context(), "test@example.com", nil, nil)
		mockStore.ExpectCreate(t.Context(), &subscription.Subscription{Email: "test@example.com"}, nil, nil)

		service := subscription.NewService(mockStore, mockLogger)

		sub := &subscription.Subscription{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := service.CreateSubscription(t.Context(), sub)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("duplicate_email", func(t *testing.T) {
		mockStore := subscriptionmock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()

		existingSub := &subscription.Subscription{
			ID:    1,
			Email: "test@example.com",
			Name:  "Existing User",
		}

		mockStore.ExpectGetByEmail(t.Context(), "test@example.com", existingSub, nil)

		service := subscription.NewService(mockStore, mockLogger)

		sub := &subscription.Subscription{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := service.CreateSubscription(t.Context(), sub)
		if !errors.Is(err, subscription.ErrEmailAlreadyExists) {
			t.Errorf("expected error %v, got %v", subscription.ErrEmailAlreadyExists, err)
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}

func TestListSubscriptions(t *testing.T) {
	mockStore := subscriptionmock.NewMockStore(t)
	mockLogger := mocklogging.NewMockLogger()
	expected := []subscription.Subscription{
		{ID: 1, Name: "Test User 1", Email: "test1@example.com"},
		{ID: 2, Name: "Test User 2", Email: "test2@example.com"},
	}

	mockStore.ExpectList(t.Context(), expected, nil)

	service := subscription.NewService(mockStore, mockLogger)

	subs, err := service.ListSubscriptions(t.Context())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !subscriptionsEqual(subs, expected) {
		t.Errorf("expected %v, got %v", expected, subs)
	}
	if err := mockStore.Verify(); err != nil {
		t.Errorf("store expectations not met: %v", err)
	}
	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}

func TestGetSubscription(t *testing.T) {
	mockStore := subscriptionmock.NewMockStore(t)
	mockLogger := mocklogging.NewMockLogger()
	service := subscription.NewService(mockStore, mockLogger)

	t.Run("existing subscription", func(t *testing.T) {
		mockStore.ExpectGetByID(t.Context(), int64(1), &subscription.Subscription{ID: 1}, nil)
		sub, err := service.GetSubscription(t.Context(), 1)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if sub == nil {
			t.Error("expected subscription, got nil")
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("non-existent subscription", func(t *testing.T) {
		mockStore.Reset()
		mockLogger.Reset()
		mockStore.ExpectGetByID(t.Context(), int64(123), nil, nil)
		mockLogger.ExpectError("failed to get subscription").WithFields(map[string]interface{}{
			"error": subscription.ErrSubscriptionNotFound,
		})

		sub, err := service.GetSubscription(t.Context(), 123)
		if !errors.Is(err, subscription.ErrSubscriptionNotFound) {
			t.Errorf("expected error %v, got %v", subscription.ErrSubscriptionNotFound, err)
		}
		if sub != nil {
			t.Errorf("expected nil subscription, got %v", sub)
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}

func TestUpdateSubscriptionStatus(t *testing.T) {
	mockStore := subscriptionmock.NewMockStore(t)
	mockLogger := mocklogging.NewMockLogger()
	service := subscription.NewService(mockStore, mockLogger)

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
				store.ExpectGetByID(t.Context(), int64(1), &subscription.Subscription{ID: 1}, nil)
				store.ExpectUpdateStatus(t.Context(), int64(1), subscription.StatusActive, nil)
			},
			wantErr: nil,
		},
		{
			name:    "invalid status",
			id:      1,
			status:  "invalid",
			setup:   func(store *subscriptionmock.MockStore) {},
			wantErr: subscription.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.Reset()
			mockLogger.Reset()
			tt.setup(mockStore)
			err := service.UpdateSubscriptionStatus(t.Context(), tt.id, tt.status)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}

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
	mockStore := subscriptionmock.NewMockStore(t)
	mockLogger := mocklogging.NewMockLogger()
	service := subscription.NewService(mockStore, mockLogger)

	t.Run("successful deletion", func(t *testing.T) {
		mockStore.ExpectGetByID(t.Context(), int64(1), &subscription.Subscription{ID: 1}, nil)
		mockStore.ExpectDelete(t.Context(), int64(1), nil)

		err := service.DeleteSubscription(t.Context(), 1)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("non-existent subscription", func(t *testing.T) {
		mockStore.Reset()
		mockLogger.Reset()
		mockStore.ExpectGetByID(t.Context(), int64(123), nil, nil)
		mockLogger.ExpectError("failed to get subscription").WithFields(map[string]interface{}{
			"error": subscription.ErrSubscriptionNotFound,
		})

		err := service.DeleteSubscription(t.Context(), 123)
		if !errors.Is(err, subscription.ErrSubscriptionNotFound) {
			t.Errorf("expected error %v, got %v", subscription.ErrSubscriptionNotFound, err)
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}

func TestGetSubscriptionByEmail(t *testing.T) {
	mockStore := subscriptionmock.NewMockStore(t)
	mockLogger := mocklogging.NewMockLogger()
	service := subscription.NewService(mockStore, mockLogger)

	t.Run("existing subscription", func(t *testing.T) {
		mockStore.ExpectGetByEmail(t.Context(), "test@example.com", &subscription.Subscription{Email: "test@example.com"}, nil)
		sub, err := service.GetSubscriptionByEmail(t.Context(), "test@example.com")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if sub == nil {
			t.Error("expected subscription, got nil")
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("non-existent subscription", func(t *testing.T) {
		mockStore.Reset()
		mockLogger.Reset()
		mockStore.ExpectGetByEmail(t.Context(), "nonexistent@example.com", nil, nil)
		mockLogger.ExpectError("failed to get subscription by email").WithFields(map[string]interface{}{
			"error": subscription.ErrSubscriptionNotFound,
		})

		sub, err := service.GetSubscriptionByEmail(t.Context(), "nonexistent@example.com")
		if !errors.Is(err, subscription.ErrSubscriptionNotFound) {
			t.Errorf("expected error %v, got %v", subscription.ErrSubscriptionNotFound, err)
		}
		if sub != nil {
			t.Errorf("expected nil subscription, got %v", sub)
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("store error", func(t *testing.T) {
		mockStore.Reset()
		mockLogger.Reset()
		storeErr := errors.New("database error")
		mockStore.ExpectGetByEmail(t.Context(), "test@example.com", nil, storeErr)
		mockLogger.ExpectError("failed to get subscription by email").WithFields(map[string]interface{}{
			"error": storeErr,
		})

		sub, err := service.GetSubscriptionByEmail(t.Context(), "test@example.com")
		if !errors.Is(err, storeErr) {
			t.Errorf("expected error %v, got %v", storeErr, err)
		}
		if sub != nil {
			t.Errorf("expected nil subscription, got %v", sub)
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("empty email", func(t *testing.T) {
		mockStore.Reset()
		mockLogger.Reset()
		sub, err := service.GetSubscriptionByEmail(t.Context(), "")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if sub != nil {
			t.Errorf("expected nil subscription, got %v", sub)
		}
		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}

// Helper function to compare subscriptions
func subscriptionEqual(a, b *subscription.Subscription) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.ID == b.ID &&
		a.Name == b.Name &&
		a.Email == b.Email &&
		a.Status == b.Status
}

// Helper function to compare subscription slices
func subscriptionsEqual(a, b []subscription.Subscription) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !subscriptionEqual(&a[i], &b[i]) {
			return false
		}
	}
	return true
}
