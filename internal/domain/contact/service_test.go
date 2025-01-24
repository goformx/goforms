package contact_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	mockstore "github.com/jonesrussell/goforms/test/mocks/contact"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestSubmitContact(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockstore.Store)
		input   *contact.Submission
		wantErr bool
	}{
		{
			name: "valid submission",
			setup: func(ms *mockstore.Store) {
				ms.On("Create", mock.Anything, mock.AnythingOfType("*contact.Submission")).Return(nil)
			},
			input: &contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			wantErr: false,
		},
		{
			name: "store error",
			setup: func(ms *mockstore.Store) {
				ms.On("Create", mock.Anything, mock.AnythingOfType("*contact.Submission")).Return(assert.AnError)
			},
			input: &contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockStore := mockstore.NewStore()
			mockLogger := mocklogging.NewMockLogger()
			mockLogger.ExpectInfo("submitting contact form")

			// Setup mock expectations
			tt.setup(mockStore)

			// Create service
			service := contact.NewService(mockStore, mockLogger)

			// Call method
			err := service.Submit(context.Background(), tt.input)

			// Assert results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			mockStore.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestListSubmissions(t *testing.T) {
	// Create test data
	submissions := []contact.Submission{
		{
			ID:      1,
			Name:    "User 1",
			Email:   "user1@example.com",
			Message: "Message 1",
		},
		{
			ID:      2,
			Name:    "User 2",
			Email:   "user2@example.com",
			Message: "Message 2",
		},
	}

	// Create mocks
	mockStore := mockstore.NewStore()
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("listing contact submissions")

	// Setup expectations
	mockStore.On("List", mock.Anything).Return(submissions, nil)

	// Create service
	service := contact.NewService(mockStore, mockLogger)

	// Call method
	result, err := service.ListSubmissions(context.Background())

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, submissions, result)

	// Verify expectations
	mockStore.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestGetSubmission(t *testing.T) {
	// Create test data
	submission := &contact.Submission{
		ID:      1,
		Name:    "Test User",
		Email:   "test@example.com",
		Message: "Test message",
	}

	// Create mocks
	mockStore := mockstore.NewStore()
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("getting contact submission")

	// Setup expectations
	mockStore.On("GetByID", mock.Anything, int64(1)).Return(submission, nil)

	// Create service
	service := contact.NewService(mockStore, mockLogger)

	// Call method
	result, err := service.GetSubmission(context.Background(), 1)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, submission, result)

	// Verify expectations
	mockStore.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestUpdateSubmissionStatus(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		status  contact.Status
		setup   func(*mockstore.Store)
		wantErr bool
	}{
		{
			name:   "valid update",
			id:     1,
			status: contact.StatusPending,
			setup: func(ms *mockstore.Store) {
				ms.On("UpdateStatus", mock.Anything, int64(1), contact.StatusPending).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "store error",
			id:     1,
			status: contact.StatusPending,
			setup: func(ms *mockstore.Store) {
				ms.On("UpdateStatus", mock.Anything, int64(1), contact.StatusPending).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockStore := mockstore.NewStore()
			mockLogger := mocklogging.NewMockLogger()
			mockLogger.ExpectInfo("updating contact submission status")

			// Setup mock expectations
			tt.setup(mockStore)

			// Create service
			service := contact.NewService(mockStore, mockLogger)

			// Call method
			err := service.UpdateSubmissionStatus(context.Background(), tt.id, tt.status)

			// Assert results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			mockStore.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
