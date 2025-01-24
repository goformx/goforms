package contact_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	storemock "github.com/jonesrussell/goforms/test/mocks/contact/store"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestSubmitContact(t *testing.T) {
	tests := []struct {
		name    string
		input   *contact.Submission
		setup   func(*storemock.MockStore, *mocklogging.MockLogger)
		wantErr bool
	}{
		{
			name: "valid_submission",
			input: &contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setup: func(ms *storemock.MockStore, ml *mocklogging.MockLogger) {
				ml.ExpectInfo("submitting contact form")
				ml.ExpectInfo("submission created",
					logging.String("email", "test@example.com"),
					logging.String("status", "pending"),
				)
				ms.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "store_error",
			input: &contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setup: func(ms *storemock.MockStore, ml *mocklogging.MockLogger) {
				ml.ExpectInfo("submitting contact form")
				ml.ExpectError("failed to create submission",
					logging.Error(assert.AnError),
					logging.String("email", "test@example.com"),
				)
				ms.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storemock.NewMockStore()
			mockLogger := mocklogging.NewMockLogger()

			tt.setup(mockStore, mockLogger)

			service := contact.NewService(mockStore, mockLogger)
			err := service.Submit(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

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
	mockStore := storemock.NewMockStore()
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("listing contact submissions",
		logging.Int("count", 2),
	)

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
	mockStore := storemock.NewMockStore()
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("getting contact submission",
		logging.Int("id", 1),
	)

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
		setup   func(*storemock.MockStore)
		wantErr bool
	}{
		{
			name:   "valid update",
			id:     1,
			status: contact.StatusPending,
			setup: func(ms *storemock.MockStore) {
				ms.On("UpdateStatus", mock.Anything, int64(1), contact.StatusPending).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "store error",
			id:     1,
			status: contact.StatusPending,
			setup: func(ms *storemock.MockStore) {
				ms.On("UpdateStatus", mock.Anything, int64(1), contact.StatusPending).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockStore := storemock.NewMockStore()
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
