package contact_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/test/mocks"
	mocklog "github.com/jonesrussell/goforms/test/mocks/logging"
)

type mockStore struct {
	mock.Mock
}

func (m *mockStore) Create(ctx context.Context, sub *contact.Submission) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *mockStore) List(ctx context.Context) ([]contact.Submission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]contact.Submission), args.Error(1)
}

func (m *mockStore) Get(ctx context.Context, id int64) (*contact.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Submission), args.Error(1)
}

func (m *mockStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	store := new(mockStore)
	logger := new(mocklog.MockLogger)
	service := contact.NewService(store, logger)
	assert.NotNil(t, service)
}

func TestService_Submit(t *testing.T) {
	tests := []struct {
		name          string
		submission    *contact.Submission
		setupMocks    func(*mockStore, *mocklog.MockLogger)
		expectedError error
	}{
		{
			name: "valid submission",
			submission: &contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupMocks: func(store *mockStore, logger *mocklog.MockLogger) {
				store.On("Create", mock.Anything, mock.MatchedBy(func(s *contact.Submission) bool {
					return s.Name == "Test User" && s.Email == "test@example.com" && s.Status == contact.StatusPending
				})).Return(nil)
				logger.On("Info", "submission created", mock.Anything).Return()
			},
			expectedError: nil,
		},
		{
			name: "missing name",
			submission: &contact.Submission{
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupMocks:    func(store *mockStore, logger *mocklog.MockLogger) {},
			expectedError: contact.ErrNameRequired,
		},
		{
			name: "missing email",
			submission: &contact.Submission{
				Name:    "Test User",
				Message: "Test message",
			},
			setupMocks:    func(store *mockStore, logger *mocklog.MockLogger) {},
			expectedError: contact.ErrEmailRequired,
		},
		{
			name: "missing message",
			submission: &contact.Submission{
				Name:  "Test User",
				Email: "test@example.com",
			},
			setupMocks:    func(store *mockStore, logger *mocklog.MockLogger) {},
			expectedError: contact.ErrMessageRequired,
		},
		{
			name: "store error",
			submission: &contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupMocks: func(store *mockStore, logger *mocklog.MockLogger) {
				store.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
				logger.On("Error", "failed to create submission", mock.Anything).Return()
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := new(mockStore)
			logger := new(mocklog.MockLogger)
			tt.setupMocks(store, logger)

			service := contact.NewService(store, logger)
			err := service.Submit(context.Background(), tt.submission)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			store.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestListSubmissions(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func(*mockStore)
		want    []contact.Submission
		wantErr bool
	}{
		{
			name: "successful list",
			setupFn: func(ms *mockStore) {
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
			setupFn: func(ms *mockStore) {
				ms.On("List", mock.Anything).Return(nil, assert.AnError)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(mockStore)
			tt.setupFn(mockStore)
			service := contact.NewService(mockStore, mocks.NewLogger())
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
		setupFn func(*mockStore)
		want    *contact.Submission
		wantErr bool
	}{
		{
			name: "existing submission",
			id:   1,
			setupFn: func(ms *mockStore) {
				ms.On("Get", mock.Anything, int64(1)).Return(&contact.Submission{
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
			setupFn: func(ms *mockStore) {
				ms.On("Get", mock.Anything, int64(999)).Return(nil, assert.AnError)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(mockStore)
			tt.setupFn(mockStore)
			service := contact.NewService(mockStore, mocks.NewLogger())
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
		setupFn func(*mockStore)
		wantErr bool
	}{
		{
			name:   "successful update",
			id:     1,
			status: contact.StatusApproved,
			setupFn: func(ms *mockStore) {
				ms.On("UpdateStatus", mock.Anything, int64(1), contact.StatusApproved).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "store error",
			id:     1,
			status: contact.StatusApproved,
			setupFn: func(ms *mockStore) {
				ms.On("UpdateStatus", mock.Anything, int64(1), contact.StatusApproved).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(mockStore)
			tt.setupFn(mockStore)
			service := contact.NewService(mockStore, mocks.NewLogger())
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
