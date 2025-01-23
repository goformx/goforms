package subscription

import (
	"context"
	"testing"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	store := NewMockStore()
	mockLogger := logger.NewMockLogger()
	service := NewService(mockLogger, store)
	assert.NotNil(t, service)
}

func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		name    string
		sub     *Subscription
		setupFn func(*MockStore)
		wantErr bool
	}{
		{
			name: "valid subscription",
			sub: &Subscription{
				Email:  "test@example.com",
				Name:   "Test User",
				Status: StatusPending,
			},
			setupFn: func(ms *MockStore) {
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, ErrSubscriptionNotFound)
				ms.On("Create", mock.Anything, mock.AnythingOfType("*subscription.Subscription")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			sub: &Subscription{
				Email:  "existing@example.com",
				Name:   "Test User",
				Status: StatusPending,
			},
			setupFn: func(ms *MockStore) {
				ms.On("GetByEmail", mock.Anything, "existing@example.com").Return(&Subscription{
					ID:     1,
					Email:  "existing@example.com",
					Status: StatusActive,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "store error",
			sub: &Subscription{
				Email:  "test@example.com",
				Name:   "Test User",
				Status: StatusPending,
			},
			setupFn: func(ms *MockStore) {
				ms.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, ErrSubscriptionNotFound)
				ms.On("Create", mock.Anything, mock.AnythingOfType("*subscription.Subscription")).Return(ErrSubscriptionNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(store)
			service := NewService(mockLogger, store)

			err := service.CreateSubscription(context.Background(), tt.sub)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			store.AssertExpectations(t)
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*MockStore)
		want    []Subscription
		wantErr bool
	}{
		{
			name: "successful list",
			setupFn: func(ms *MockStore) {
				ms.On("List", mock.Anything).Return([]Subscription{
					{ID: 1, Email: "test1@example.com", Status: StatusActive},
					{ID: 2, Email: "test2@example.com", Status: StatusPending},
				}, nil)
			},
			want: []Subscription{
				{ID: 1, Email: "test1@example.com", Status: StatusActive},
				{ID: 2, Email: "test2@example.com", Status: StatusPending},
			},
			wantErr: false,
		},
		{
			name: "store error",
			setupFn: func(ms *MockStore) {
				ms.On("List", mock.Anything).Return(nil, ErrSubscriptionNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(store)
			service := NewService(mockLogger, store)

			got, err := service.ListSubscriptions(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			store.AssertExpectations(t)
		})
	}
}

func TestGetSubscription(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		setupFn func(*MockStore)
		want    *Subscription
		wantErr bool
	}{
		{
			name: "existing subscription",
			id:   1,
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&Subscription{
					ID:     1,
					Email:  "test@example.com",
					Status: StatusActive,
				}, nil)
			},
			want: &Subscription{
				ID:     1,
				Email:  "test@example.com",
				Status: StatusActive,
			},
			wantErr: false,
		},
		{
			name: "non-existent subscription",
			id:   999,
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(999)).Return(nil, ErrSubscriptionNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(store)
			service := NewService(mockLogger, store)

			got, err := service.GetSubscription(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			store.AssertExpectations(t)
		})
	}
}

func TestUpdateSubscriptionStatus(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		status  Status
		setupFn func(*MockStore)
		wantErr bool
	}{
		{
			name:   "valid update",
			id:     1,
			status: StatusActive,
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&Subscription{
					ID:     1,
					Email:  "test@example.com",
					Status: StatusPending,
				}, nil)
				ms.On("UpdateStatus", mock.Anything, int64(1), StatusActive).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "non-existent subscription",
			id:     999,
			status: StatusActive,
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(999)).Return(nil, ErrSubscriptionNotFound)
			},
			wantErr: true,
		},
		{
			name:   "invalid status",
			id:     1,
			status: "invalid",
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&Subscription{
					ID:     1,
					Email:  "test@example.com",
					Status: StatusPending,
				}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(store)
			service := NewService(mockLogger, store)

			err := service.UpdateSubscriptionStatus(context.Background(), tt.id, tt.status)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			store.AssertExpectations(t)
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		setupFn func(*MockStore)
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   1,
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&Subscription{
					ID:     1,
					Email:  "test@example.com",
					Status: StatusActive,
				}, nil)
				ms.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "non-existent subscription",
			id:   999,
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(999)).Return(nil, ErrSubscriptionNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(store)
			service := NewService(mockLogger, store)

			err := service.DeleteSubscription(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			store.AssertExpectations(t)
		})
	}
}
