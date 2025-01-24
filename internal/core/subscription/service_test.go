package subscription_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/core/subscription"
	"github.com/jonesrussell/goforms/internal/logger"
	storemock "github.com/jonesrussell/goforms/test/mocks/store/subscription"
)

func TestNewService(t *testing.T) {
	store := storemock.NewMockStore()
	mockLogger := logger.NewMockLogger()
	service := subscription.NewService(mockLogger, store)
	assert.NotNil(t, service)
}

func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		name    string
		sub     *subscription.Subscription
		setupFn func(*storemock.MockStore)
		wantErr bool
	}{
		{
			name: "successful create",
			sub: &subscription.Subscription{
				Email:  "test@example.com",
				Name:   "Test User",
				Status: subscription.StatusPending,
			},
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, subscription.ErrSubscriptionNotFound)
				ms.On("Create", mock.Anything, mock.AnythingOfType("*subscription.Subscription")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			sub: &subscription.Subscription{
				Email:  "existing@example.com",
				Name:   "Test User",
				Status: subscription.StatusPending,
			},
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByEmail", mock.Anything, "existing@example.com").Return(&subscription.Subscription{
					ID:     1,
					Email:  "existing@example.com",
					Status: subscription.StatusActive,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "store error",
			sub: &subscription.Subscription{
				Email:  "test@example.com",
				Name:   "Test User",
				Status: subscription.StatusPending,
			},
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, subscription.ErrSubscriptionNotFound)
				ms.On("Create", mock.Anything, mock.AnythingOfType("*subscription.Subscription")).Return(subscription.ErrSubscriptionNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			tt.setupFn(mockStore)
			service := subscription.NewService(logger.NewMockLogger(), mockStore)

			err := service.CreateSubscription(context.Background(), tt.sub)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*storemock.MockStore)
		want    []subscription.Subscription
		wantErr bool
	}{
		{
			name: "successful list",
			setupFn: func(ms *storemock.MockStore) {
				ms.On("List", mock.Anything).Return([]subscription.Subscription{
					{ID: 1, Email: "test1@example.com", Status: subscription.StatusActive},
					{ID: 2, Email: "test2@example.com", Status: subscription.StatusPending},
				}, nil)
			},
			want: []subscription.Subscription{
				{ID: 1, Email: "test1@example.com", Status: subscription.StatusActive},
				{ID: 2, Email: "test2@example.com", Status: subscription.StatusPending},
			},
			wantErr: false,
		},
		{
			name: "store error",
			setupFn: func(ms *storemock.MockStore) {
				ms.On("List", mock.Anything).Return(nil, subscription.ErrSubscriptionNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(mockStore)
			service := subscription.NewService(mockLogger, mockStore)

			got, err := service.ListSubscriptions(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetSubscription(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		setupFn func(*storemock.MockStore)
		want    *subscription.Subscription
		wantErr bool
	}{
		{
			name: "existing subscription",
			id:   1,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&subscription.Subscription{
					ID:     1,
					Email:  "test@example.com",
					Status: subscription.StatusActive,
				}, nil)
			},
			want: &subscription.Subscription{
				ID:     1,
				Email:  "test@example.com",
				Status: subscription.StatusActive,
			},
			wantErr: false,
		},
		{
			name: "non-existent subscription",
			id:   999,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(999)).Return(nil, subscription.ErrSubscriptionNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(mockStore)
			service := subscription.NewService(mockLogger, mockStore)

			got, err := service.GetSubscription(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUpdateSubscriptionStatus(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		status  subscription.Status
		setupFn func(*storemock.MockStore)
		wantErr bool
	}{
		{
			name:   "valid update",
			id:     1,
			status: subscription.StatusActive,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&subscription.Subscription{
					ID:     1,
					Email:  "test@example.com",
					Status: subscription.StatusPending,
				}, nil)
				ms.On("UpdateStatus", mock.Anything, int64(1), subscription.StatusActive).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "non-existent subscription",
			id:     999,
			status: subscription.StatusActive,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(999)).Return(nil, subscription.ErrSubscriptionNotFound)
			},
			wantErr: true,
		},
		{
			name:   "invalid status",
			id:     1,
			status: "invalid",
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&subscription.Subscription{
					ID:     1,
					Email:  "test@example.com",
					Status: subscription.StatusPending,
				}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(mockStore)
			service := subscription.NewService(mockLogger, mockStore)

			err := service.UpdateSubscriptionStatus(context.Background(), tt.id, tt.status)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		setupFn func(*storemock.MockStore)
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   1,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&subscription.Subscription{
					ID:     1,
					Email:  "test@example.com",
					Status: subscription.StatusActive,
				}, nil)
				ms.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "non-existent subscription",
			id:   999,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(999)).Return(nil, subscription.ErrSubscriptionNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(mockStore)
			service := subscription.NewService(mockLogger, mockStore)

			err := service.DeleteSubscription(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetSubscriptionByEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		setup   func(*storemock.MockStore)
		want    *subscription.Subscription
		wantErr error
	}{
		{
			name:  "existing subscription",
			email: "test@example.com",
			setup: func(ms *storemock.MockStore) {
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
			setup: func(ms *storemock.MockStore) {
				ms.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, subscription.ErrSubscriptionNotFound)
			},
			want:    nil,
			wantErr: errors.New("failed to get subscription by email: subscription not found"),
		},
		{
			name:  "store error",
			email: "test@example.com",
			setup: func(ms *storemock.MockStore) {
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("database error"))
			},
			want:    nil,
			wantErr: errors.New("failed to get subscription by email: database error"),
		},
		{
			name:  "empty email",
			email: "",
			setup: func(ms *storemock.MockStore) {
				// No mock setup needed as it should fail before store call
			},
			want:    nil,
			wantErr: errors.New("invalid input: email is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock store for each test
			mockStore := storemock.NewMockStore()
			mockLogger := logger.NewMockLogger()

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
