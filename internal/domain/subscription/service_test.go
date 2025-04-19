package subscription_test

import (
	"errors"
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	subscriptionmock "github.com/jonesrussell/goforms/test/mocks/store/subscription"
	"github.com/stretchr/testify/require"
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

		createErr := service.CreateSubscription(t.Context(), sub)
		if createErr != nil {
			t.Fatalf("unexpected error: %v", createErr)
		}

		if verifyErr := mockStore.Verify(); verifyErr != nil {
			t.Fatalf("mock store verification failed: %v", verifyErr)
		}

		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("mock logger verification failed: %v", verifyErr)
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

		createErr := service.CreateSubscription(t.Context(), sub)
		if !errors.Is(createErr, subscription.ErrEmailAlreadyExists) {
			t.Errorf("expected error %v, got %v", subscription.ErrEmailAlreadyExists, createErr)
		}
		if verifyErr := mockStore.Verify(); verifyErr != nil {
			t.Fatalf("mock store verification failed: %v", verifyErr)
		}
		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("mock logger verification failed: %v", verifyErr)
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

	subs, listErr := service.ListSubscriptions(t.Context())
	if listErr != nil {
		t.Fatalf("unexpected error: %v", listErr)
	}
	if !subscriptionsEqual(subs, expected) {
		t.Errorf("expected %v, got %v", expected, subs)
	}
	if verifyErr := mockStore.Verify(); verifyErr != nil {
		t.Fatalf("mock store verification failed: %v", verifyErr)
	}
	if verifyErr := mockLogger.Verify(); verifyErr != nil {
		t.Fatalf("mock logger verification failed: %v", verifyErr)
	}
}

func TestGetSubscription(t *testing.T) {
	mockStore := subscriptionmock.NewMockStore(t)
	mockLogger := mocklogging.NewMockLogger()
	service := subscription.NewService(mockStore, mockLogger)

	t.Run("existing subscription", func(t *testing.T) {
		mockStore.ExpectGetByID(t.Context(), int64(1), &subscription.Subscription{ID: 1}, nil)
		sub, getErr := service.GetSubscription(t.Context(), 1)
		if getErr != nil {
			t.Fatalf("unexpected error: %v", getErr)
		}
		if sub == nil {
			t.Error("expected subscription, got nil")
		}
		if verifyErr := mockStore.Verify(); verifyErr != nil {
			t.Fatalf("mock store verification failed: %v", verifyErr)
		}
		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("mock logger verification failed: %v", verifyErr)
		}
	})

	t.Run("non-existent subscription", func(t *testing.T) {
		mockStore.Reset()
		mockLogger.Reset()
		mockStore.ExpectGetByID(t.Context(), int64(123), nil, nil)
		mockLogger.ExpectError("failed to get subscription").WithFields(map[string]any{
			"error": subscription.ErrSubscriptionNotFound,
		})

		sub, getErr := service.GetSubscription(t.Context(), 123)
		if !errors.Is(getErr, subscription.ErrSubscriptionNotFound) {
			t.Errorf("expected error %v, got %v", subscription.ErrSubscriptionNotFound, getErr)
		}
		if sub != nil {
			t.Errorf("expected nil subscription, got %v", sub)
		}
		if verifyErr := mockStore.Verify(); verifyErr != nil {
			t.Fatalf("mock store verification failed: %v", verifyErr)
		}
		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("mock logger verification failed: %v", verifyErr)
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
			updateErr := service.UpdateSubscriptionStatus(t.Context(), tt.id, tt.status)

			if !errors.Is(updateErr, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, updateErr)
			}

			if verifyErr := mockStore.Verify(); verifyErr != nil {
				t.Fatalf("mock store verification failed: %v", verifyErr)
			}
			if verifyErr := mockLogger.Verify(); verifyErr != nil {
				t.Fatalf("mock logger verification failed: %v", verifyErr)
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

		deleteErr := service.DeleteSubscription(t.Context(), 1)
		if deleteErr != nil {
			t.Fatalf("unexpected error: %v", deleteErr)
		}
		if verifyErr := mockStore.Verify(); verifyErr != nil {
			t.Fatalf("mock store verification failed: %v", verifyErr)
		}
		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("mock logger verification failed: %v", verifyErr)
		}
	})

	t.Run("non-existent subscription", func(t *testing.T) {
		mockStore.Reset()
		mockLogger.Reset()
		mockStore.ExpectGetByID(t.Context(), int64(123), nil, nil)
		mockLogger.ExpectError("failed to get subscription").WithFields(map[string]any{
			"error": subscription.ErrSubscriptionNotFound,
		})

		deleteErr := service.DeleteSubscription(t.Context(), 123)
		if !errors.Is(deleteErr, subscription.ErrSubscriptionNotFound) {
			t.Errorf("expected error %v, got %v", subscription.ErrSubscriptionNotFound, deleteErr)
		}
		if verifyErr := mockStore.Verify(); verifyErr != nil {
			t.Fatalf("mock store verification failed: %v", verifyErr)
		}
		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("mock logger verification failed: %v", verifyErr)
		}
	})
}

func TestGetSubscriptionByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		expectedSub   *subscription.Subscription
		expectedError error
		setupMock     func(*subscriptionmock.MockStore)
	}{
		{
			name:          "subscription found",
			email:         "test@example.com",
			expectedSub:   &subscription.Subscription{Email: "test@example.com"},
			expectedError: nil,
			setupMock: func(store *subscriptionmock.MockStore) {
				store.ExpectGetByEmail(t.Context(), "test@example.com", &subscription.Subscription{Email: "test@example.com"}, nil)
			},
		},
		{
			name:          "subscription not found",
			email:         "notfound@example.com",
			expectedSub:   nil,
			expectedError: subscription.ErrSubscriptionNotFound,
			setupMock: func(store *subscriptionmock.MockStore) {
				store.ExpectGetByEmail(t.Context(), "notfound@example.com", nil, subscription.ErrSubscriptionNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockStore := subscriptionmock.NewMockStore(t)
			tt.setupMock(mockStore)
			service := subscription.NewService(mockStore, mocklogging.NewMockLogger())

			// Execute
			sub, err := service.GetSubscriptionByEmail(t.Context(), tt.email)

			// Verify
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedSub, sub)
			}
			require.NoError(t, mockStore.Verify())
		})
	}
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
