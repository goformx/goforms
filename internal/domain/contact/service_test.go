package contact_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	contactmock "github.com/jonesrussell/goforms/test/mocks/contact/store"
	loggingmock "github.com/jonesrussell/goforms/test/mocks/logging"
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
				ms.CreateFunc = func(ctx context.Context, sub *contact.Submission) error {
					return nil
				}
				ml.ExpectInfo("submission created").WithFields(map[string]interface{}{
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
				ms.CreateFunc = func(ctx context.Context, sub *contact.Submission) error {
					return errTest
				}
				ml.ExpectError("failed to create submission").WithFields(map[string]interface{}{
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
			err := svc.Submit(context.Background(), tt.input)

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
				ms.ListFunc = func(ctx context.Context) ([]contact.Submission, error) {
					return []contact.Submission{{ID: 1}}, nil
				}
			},
			want:    []contact.Submission{{ID: 1}},
			wantErr: false,
		},
		{
			name: "store_error",
			setup: func(ms *contactmock.MockStore, ml *loggingmock.MockLogger) {
				ms.ListFunc = func(ctx context.Context) ([]contact.Submission, error) {
					return nil, errTest
				}
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
			got, err := svc.ListSubmissions(context.Background())

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
				ms.GetFunc = func(ctx context.Context, id int64) (*contact.Submission, error) {
					return &contact.Submission{ID: id}, nil
				}
			},
			id:      1,
			want:    &contact.Submission{ID: 1},
			wantErr: false,
		},
		{
			name: "store_error",
			setup: func(ms *contactmock.MockStore, ml *loggingmock.MockLogger) {
				ms.GetFunc = func(ctx context.Context, id int64) (*contact.Submission, error) {
					return nil, errTest
				}
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
			got, err := svc.GetSubmission(context.Background(), tt.id)

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
