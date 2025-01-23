package contact

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
	service := NewService(store, mockLogger)
	assert.NotNil(t, service)
}

func TestCreateSubmission(t *testing.T) {
	tests := []struct {
		name       string
		submission *Submission
		setupFn    func(*MockStore)
		wantErr    bool
	}{
		{
			name: "successful creation",
			submission: &Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupFn: func(ms *MockStore) {
				ms.On("Create", mock.Anything, mock.MatchedBy(func(s *Submission) bool {
					return s.Status == StatusPending &&
						s.Name == "Test User" &&
						s.Email == "test@example.com"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "store error",
			submission: &Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupFn: func(ms *MockStore) {
				ms.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(store)

			service := NewService(store, mockLogger)
			err := service.CreateSubmission(context.Background(), tt.submission)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			store.AssertExpectations(t)
		})
	}
}

func TestListSubmissions(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*MockStore)
		want    []Submission
		wantErr bool
	}{
		{
			name: "successful list",
			setupFn: func(ms *MockStore) {
				ms.On("List", mock.Anything).Return([]Submission{
					{
						ID:      1,
						Name:    "Test User",
						Email:   "test@example.com",
						Message: "Test message",
						Status:  StatusPending,
					},
				}, nil)
			},
			want: []Submission{
				{
					ID:      1,
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
					Status:  StatusPending,
				},
			},
			wantErr: false,
		},
		{
			name: "store error",
			setupFn: func(ms *MockStore) {
				ms.On("List", mock.Anything).Return(nil, assert.AnError)
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

			service := NewService(store, mockLogger)
			got, err := service.ListSubmissions(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			store.AssertExpectations(t)
		})
	}
}

func TestGetSubmission(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		setupFn func(*MockStore)
		want    *Submission
		wantErr bool
	}{
		{
			name: "existing submission",
			id:   1,
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&Submission{
					ID:      1,
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
					Status:  StatusPending,
				}, nil)
			},
			want: &Submission{
				ID:      1,
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
				Status:  StatusPending,
			},
			wantErr: false,
		},
		{
			name: "non-existent submission",
			id:   999,
			setupFn: func(ms *MockStore) {
				ms.On("GetByID", mock.Anything, int64(999)).Return(nil, assert.AnError)
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

			service := NewService(store, mockLogger)
			got, err := service.GetSubmission(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			store.AssertExpectations(t)
		})
	}
}

func TestUpdateSubmissionStatus(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		status  Status
		setupFn func(*MockStore)
		wantErr bool
	}{
		{
			name:   "successful update",
			id:     1,
			status: StatusApproved,
			setupFn: func(ms *MockStore) {
				ms.On("UpdateStatus", mock.Anything, int64(1), StatusApproved).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "store error",
			id:     1,
			status: StatusApproved,
			setupFn: func(ms *MockStore) {
				ms.On("UpdateStatus", mock.Anything, int64(1), StatusApproved).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockStore()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(store)

			service := NewService(store, mockLogger)
			err := service.UpdateSubmissionStatus(context.Background(), tt.id, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			store.AssertExpectations(t)
		})
	}
}
