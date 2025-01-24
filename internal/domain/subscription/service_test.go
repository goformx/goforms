package subscription_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/test/mocks"
)

func TestNewService(t *testing.T) {
	mockStore := mocks.NewSubscriptionStore()
	mockLogger := mocks.NewLogger()
	service := subscription.NewService(mockLogger, mockStore)
	assert.NotNil(t, service)
}

func TestCreateSubscription(t *testing.T) {
	// Create mock store
	mockStore := mocks.NewSubscriptionStore()
	mockLogger := mocks.NewLogger()

	// Create service with mocks
	service := subscription.NewService(mockLogger, mockStore)

	// Test cases
	tests := []struct {
		name    string
		sub     *subscription.Subscription
		setup   func()
		wantErr error
	}{
		{
			name: "valid subscription",
			sub: &subscription.Subscription{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setup: func() {
				mockStore.On("GetByEmail", mock.Anything, "test@example.com").
					Return(nil, subscription.ErrSubscriptionNotFound)
				mockStore.On("Create", mock.Anything, mock.AnythingOfType("*subscription.Subscription")).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "duplicate email",
			sub: &subscription.Subscription{
				Email: "existing@example.com",
				Name:  "Existing User",
			},
			setup: func() {
				mockStore.On("GetByEmail", mock.Anything, "existing@example.com").
					Return(&subscription.Subscription{}, nil)
			},
			wantErr: subscription.ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.setup()

			// Call service method
			err := service.CreateSubscription(context.Background(), tt.sub)

			// Assert error
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockStore.AssertExpectations(t)
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	// Create mock store
	mockStore := mocks.NewSubscriptionStore()
	mockLogger := mocks.NewLogger()

	// Create service with mocks
	service := subscription.NewService(mockLogger, mockStore)

	// Setup test data
	expectedSubs := []subscription.Subscription{
		{ID: 1, Email: "test1@example.com"},
		{ID: 2, Email: "test2@example.com"},
	}

	// Setup mock expectations
	mockStore.On("List", mock.Anything).Return(expectedSubs, nil)

	// Call service method
	subs, err := service.ListSubscriptions(context.Background())

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedSubs, subs)

	// Verify mock expectations
	mockStore.AssertExpectations(t)
}

func TestGetSubscription(t *testing.T) {
	// Create mock store
	mockStore := mocks.NewSubscriptionStore()
	mockLogger := mocks.NewLogger()

	// Create service with mocks
	service := subscription.NewService(mockLogger, mockStore)

	// Test cases
	tests := []struct {
		name    string
		id      int64
		setup   func()
		want    *subscription.Subscription
		wantErr error
	}{
		{
			name: "existing subscription",
			id:   1,
			setup: func() {
				mockStore.On("GetByID", mock.Anything, int64(1)).
					Return(&subscription.Subscription{ID: 1, Email: "test@example.com"}, nil)
			},
			want: &subscription.Subscription{ID: 1, Email: "test@example.com"},
		},
		{
			name: "non-existent subscription",
			id:   999,
			setup: func() {
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
			got, err := service.GetSubscription(context.Background(), tt.id)

			// Assert results
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			// Verify mock expectations
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUpdateSubscriptionStatus(t *testing.T) {
	// Create mock store
	mockStore := &mocks.SubscriptionStore{}
	mockLogger := mocks.NewLogger()

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
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	// Create mock store
	mockStore := &mocks.SubscriptionStore{}
	mockLogger := mocks.NewLogger()

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
				mockStore.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
		},
		{
			name: "non-existent subscription",
			id:   999,
			setup: func() {
				mockStore.On("Delete", mock.Anything, int64(999)).
					Return(subscription.ErrSubscriptionNotFound)
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
		})
	}
}

func TestGetSubscriptionByEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		setup   func(*mocks.SubscriptionStore)
		want    *subscription.Subscription
		wantErr error
	}{
		{
			name:  "existing subscription",
			email: "test@example.com",
			setup: func(ms *mocks.SubscriptionStore) {
				sub := &subscription.Subscription{
					ID:     1,
					Email:  "test@example.com",
					Name:   "Test User",
					Status: subscription.StatusActive,
				}
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(sub, nil)
			},
			want: &subscription.Subscription{
				ID:     1,
				Email:  "test@example.com",
				Name:   "Test User",
				Status: subscription.StatusActive,
			},
			wantErr: nil,
		},
		{
			name:  "non-existent subscription",
			email: "nonexistent@example.com",
			setup: func(ms *mocks.SubscriptionStore) {
				ms.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, subscription.ErrSubscriptionNotFound)
			},
			want:    nil,
			wantErr: errors.New("failed to get subscription by email: subscription not found"),
		},
		{
			name:  "store error",
			email: "test@example.com",
			setup: func(ms *mocks.SubscriptionStore) {
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("database error"))
			},
			want:    nil,
			wantErr: errors.New("failed to get subscription by email: database error"),
		},
		{
			name:  "empty email",
			email: "",
			setup: func(ms *mocks.SubscriptionStore) {
				// No mock setup needed as it should fail before store call
			},
			want:    nil,
			wantErr: errors.New("invalid input: email is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock store for each test
			mockStore := mocks.NewSubscriptionStore()
			mockLogger := mocks.NewLogger()

			// Setup the mock expectations
			tt.setup(mockStore)

			// Create the service with the mock store
			service := subscription.NewService(mockLogger, mockStore)

			// Call the method being tested
			got, err := service.GetSubscriptionByEmail(context.Background(), tt.email)

			// Assert the results
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)

			// Verify all expectations were met
			mockStore.AssertExpectations(t)
		})
	}
}
