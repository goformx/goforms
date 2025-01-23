package contact_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/jonesrussell/goforms/internal/logger"
	storemock "github.com/jonesrussell/goforms/test/mocks/store/contact"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	store := storemock.NewMockStore()
	mockLogger := logger.NewMockLogger()
	service := contact.NewService(store, mockLogger)
	assert.NotNil(t, service)
}

func TestCreateSubmission(t *testing.T) {
	tests := []struct {
		name    string
		sub     *contact.Submission
		setupFn func(*storemock.MockStore)
		wantErr bool
	}{
		{
			name: "successful create",
			sub: &contact.Submission{
				Email:   "test@example.com",
				Name:    "Test User",
				Message: "Test message",
				Status:  contact.StatusPending,
			},
			setupFn: func(ms *storemock.MockStore) {
				ms.On("Create", mock.Anything, mock.AnythingOfType("*contact.Submission")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "store error",
			sub: &contact.Submission{
				Email:   "test@example.com",
				Name:    "Test User",
				Message: "Test message",
				Status:  contact.StatusPending,
			},
			setupFn: func(ms *storemock.MockStore) {
				ms.On("Create", mock.Anything, mock.AnythingOfType("*contact.Submission")).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			tt.setupFn(mockStore)
			service := contact.NewService(mockStore, logger.NewMockLogger())

			err := service.CreateSubmission(context.Background(), tt.sub)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestListSubmissions(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*storemock.MockStore)
		want    []contact.Submission
		wantErr bool
	}{
		{
			name: "successful list",
			setupFn: func(ms *storemock.MockStore) {
				ms.On("List", mock.Anything).Return([]contact.Submission{
					{
						ID:      1,
						Name:    "Test User",
						Email:   "test@example.com",
						Message: "Test message",
						Status:  contact.StatusPending,
					},
				}, nil)
			},
			want: []contact.Submission{
				{
					ID:      1,
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
					Status:  contact.StatusPending,
				},
			},
			wantErr: false,
		},
		{
			name: "store error",
			setupFn: func(ms *storemock.MockStore) {
				ms.On("List", mock.Anything).Return(nil, assert.AnError)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			tt.setupFn(mockStore)
			service := contact.NewService(mockStore, logger.NewMockLogger())
			got, err := service.ListSubmissions(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetSubmission(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		setupFn func(*storemock.MockStore)
		want    *contact.Submission
		wantErr bool
	}{
		{
			name: "existing submission",
			id:   1,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(1)).Return(&contact.Submission{
					ID:      1,
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
					Status:  contact.StatusPending,
				}, nil)
			},
			want: &contact.Submission{
				ID:      1,
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
				Status:  contact.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "non-existent submission",
			id:   999,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("GetByID", mock.Anything, int64(999)).Return(nil, assert.AnError)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			tt.setupFn(mockStore)
			service := contact.NewService(mockStore, logger.NewMockLogger())
			got, err := service.GetSubmission(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUpdateSubmissionStatus(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		status  contact.Status
		setupFn func(*storemock.MockStore)
		wantErr bool
	}{
		{
			name:   "successful update",
			id:     1,
			status: contact.StatusApproved,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("UpdateStatus", mock.Anything, int64(1), contact.StatusApproved).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "store error",
			id:     1,
			status: contact.StatusApproved,
			setupFn: func(ms *storemock.MockStore) {
				ms.On("UpdateStatus", mock.Anything, int64(1), contact.StatusApproved).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			tt.setupFn(mockStore)
			service := contact.NewService(mockStore, logger.NewMockLogger())
			err := service.UpdateSubmissionStatus(context.Background(), tt.id, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockStore.AssertExpectations(t)
		})
	}
}
