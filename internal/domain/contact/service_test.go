package contact_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	contactmock "github.com/jonesrussell/goforms/test/mocks/contact/store"
	loggingmock "github.com/jonesrussell/goforms/test/mocks/logging"
	"github.com/stretchr/testify/mock"
)

var errTest = errors.New("test error")

func TestSubmitContact(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*contactmock.MockStore, *loggingmock.MockLogger)
		input   *contact.Submission
		wantErr bool
	}{
		{
			name: "valid_submission",
			setup: func(ms *contactmock.MockStore, ml *loggingmock.MockLogger) {
				ms.On("Create", mock.Anything, mock.AnythingOfType("*contact.Submission")).Return(nil)
				ml.ExpectInfo("submission created").WithFields(map[string]any{
					"email":  "test@example.com",
					"status": string(contact.StatusPending),
				})
			},
			input: &contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			wantErr: false,
		},
		{
			name: "store_error",
			setup: func(ms *contactmock.MockStore, ml *loggingmock.MockLogger) {
				ms.On("Create", mock.Anything, mock.AnythingOfType("*contact.Submission")).Return(errTest)
				ml.ExpectError("failed to create submission").WithFields(map[string]any{
					"error": errTest,
					"email": "test@example.com",
				})
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
			mockStore := contactmock.NewMockStore()
			mockLogger := loggingmock.NewMockLogger()
			tt.setup(mockStore, mockLogger)

			svc := contact.NewService(mockStore, mockLogger)
			err := svc.Submit(t.Context(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Submit() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mockLogger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
		})
	}
}

func TestListSubmissions(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*contactmock.MockStore, *loggingmock.MockLogger)
		want    []contact.Submission
		wantErr bool
	}{
		{
			name: "success",
			setup: func(ms *contactmock.MockStore, ml *loggingmock.MockLogger) {
				ms.On("List", mock.Anything).Return([]contact.Submission{{ID: 1}}, nil)
			},
			want:    []contact.Submission{{ID: 1}},
			wantErr: false,
		},
		{
			name: "store_error",
			setup: func(ms *contactmock.MockStore, ml *loggingmock.MockLogger) {
				ms.On("List", mock.Anything).Return(nil, errTest)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := contactmock.NewMockStore()
			mockLogger := loggingmock.NewMockLogger()
			tt.setup(mockStore, mockLogger)

			svc := contact.NewService(mockStore, mockLogger)
			got, err := svc.ListSubmissions(t.Context())

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !submissionsEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}

			if err := mockLogger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
		})
	}
}

func TestGetSubmission(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*contactmock.MockStore, *loggingmock.MockLogger)
		id      int64
		want    *contact.Submission
		wantErr bool
	}{
		{
			name: "success",
			setup: func(ms *contactmock.MockStore, ml *loggingmock.MockLogger) {
				ms.On("Get", mock.Anything, int64(1)).Return(&contact.Submission{ID: 1}, nil)
			},
			id:      1,
			want:    &contact.Submission{ID: 1},
			wantErr: false,
		},
		{
			name: "store_error",
			setup: func(ms *contactmock.MockStore, ml *loggingmock.MockLogger) {
				ms.On("Get", mock.Anything, int64(1)).Return(nil, errTest)
			},
			id:      1,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := contactmock.NewMockStore()
			mockLogger := loggingmock.NewMockLogger()
			tt.setup(mockStore, mockLogger)

			svc := contact.NewService(mockStore, mockLogger)
			got, err := svc.GetSubmission(t.Context(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !submissionEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}

			if err := mockLogger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
		})
	}
}

// Helper function to compare submissions
func submissionEqual(a, b *contact.Submission) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.ID == b.ID &&
		a.Name == b.Name &&
		a.Email == b.Email &&
		a.Message == b.Message &&
		a.Status == b.Status
}

// Helper function to compare submission slices
func submissionsEqual(a, b []contact.Submission) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !submissionEqual(&a[i], &b[i]) {
			return false
		}
	}
	return true
}

func TestContactService(t *testing.T) {
	t.Run("Submit", func(t *testing.T) {
		mockLogger := loggingmock.NewMockLogger()
		mockStore := contactmock.NewMockStore()
		service := contact.NewService(mockStore, mockLogger)

		submission := &contact.Submission{
			Email:   "test@example.com",
			Message: "Test message",
		}

		mockStore.On("Create", mock.Anything, submission).Return(nil)
		mockLogger.ExpectInfo("contact submission created")

		if err := service.Submit(context.Background(), submission); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("unexpected error: %v", verifyErr)
		}
	})

	t.Run("ListSubmissions", func(t *testing.T) {
		mockLogger := loggingmock.NewMockLogger()
		mockStore := contactmock.NewMockStore()
		service := contact.NewService(mockStore, mockLogger)

		submissions := []contact.Submission{
			{
				Email:   "test1@example.com",
				Message: "Test message 1",
			},
			{
				Email:   "test2@example.com",
				Message: "Test message 2",
			},
		}

		mockStore.On("List", mock.Anything).Return(submissions, nil)
		mockLogger.ExpectInfo("contact submissions listed")

		result, err := service.ListSubmissions(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != len(submissions) {
			t.Fatalf("expected %d submissions, got %d", len(submissions), len(result))
		}

		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("unexpected error: %v", verifyErr)
		}
	})

	t.Run("GetSubmission", func(t *testing.T) {
		mockLogger := loggingmock.NewMockLogger()
		mockStore := contactmock.NewMockStore()
		service := contact.NewService(mockStore, mockLogger)

		submission := &contact.Submission{
			ID:      1,
			Email:   "test@example.com",
			Message: "Test message",
		}

		mockStore.On("Get", mock.Anything, int64(1)).Return(submission, nil)
		mockLogger.ExpectInfo("contact submission retrieved")

		result, err := service.GetSubmission(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.ID != submission.ID {
			t.Fatalf("expected submission ID %d, got %d", submission.ID, result.ID)
		}

		if verifyErr := mockLogger.Verify(); verifyErr != nil {
			t.Fatalf("unexpected error: %v", verifyErr)
		}
	})
}
