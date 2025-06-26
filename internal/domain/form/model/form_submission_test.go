package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

func TestFormSubmission_Validate(t *testing.T) {
	tests := []struct {
		name        string
		submission  *model.FormSubmission
		wantErr     bool
		errContains string
	}{
		{
			name: "valid submission",
			submission: &model.FormSubmission{
				FormID:      "form123",
				Data:        model.JSON{"name": "John Doe", "email": "john@example.com"},
				SubmittedAt: time.Now(),
				Status:      model.SubmissionStatusPending,
				Metadata:    model.JSON{"source": "web"},
			},
			wantErr: false,
		},
		{
			name: "missing form ID",
			submission: &model.FormSubmission{
				Data:        model.JSON{"name": "John Doe"},
				SubmittedAt: time.Now(),
				Status:      model.SubmissionStatusPending,
			},
			wantErr:     true,
			errContains: "form ID is required",
		},
		{
			name: "missing submission data",
			submission: &model.FormSubmission{
				FormID:      "form123",
				SubmittedAt: time.Now(),
				Status:      model.SubmissionStatusPending,
			},
			wantErr:     true,
			errContains: "submission data is required",
		},
		{
			name: "empty submission data",
			submission: &model.FormSubmission{
				FormID:      "form123",
				Data:        model.JSON{},
				SubmittedAt: time.Now(),
				Status:      model.SubmissionStatusPending,
			},
			wantErr: false, // Empty data is valid
		},
		{
			name: "missing status defaults to pending",
			submission: &model.FormSubmission{
				FormID:      "form123",
				Data:        model.JSON{"name": "John Doe"},
				SubmittedAt: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.submission.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFormSubmission_Sanitize(t *testing.T) {
	sanitizer := sanitization.NewService()

	tests := []struct {
		name     string
		input    *model.FormSubmission
		expected *model.FormSubmission
	}{
		{
			name: "sanitize XSS in data",
			input: &model.FormSubmission{
				FormID: "form123",
				Data: model.JSON{
					"name":  "<script>alert('xss')</script>John",
					"email": "john@example.com",
				},
				Metadata: model.JSON{
					"comment": "<img src=x onerror=alert('xss')>",
				},
			},
			expected: &model.FormSubmission{
				FormID: "form123",
				Data: model.JSON{
					"name":  ">alert('xss')</John",
					"email": "john@example.com",
				},
				Metadata: model.JSON{
					"comment": "<img src=x alert('xss')>",
				},
			},
		},
		{
			name: "sanitize non-string values",
			input: &model.FormSubmission{
				FormID: "form123",
				Data: model.JSON{
					"age":    25,
					"active": true,
					"tags":   []string{"tag1", "tag2"},
				},
			},
			expected: &model.FormSubmission{
				FormID: "form123",
				Data: model.JSON{
					"age":    25,
					"active": true,
					"tags":   []string{"tag1", "tag2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Sanitize(sanitizer)
			require.Equal(t, tt.expected.Data, tt.input.Data)
			require.Equal(t, tt.expected.Metadata, tt.input.Metadata)
		})
	}
}

func TestFormSubmission_StatusMethods(t *testing.T) {
	tests := []struct {
		name         string
		status       model.SubmissionStatus
		isCompleted  bool
		isFailed     bool
		isPending    bool
		isProcessing bool
	}{
		{
			name:         "completed status",
			status:       model.SubmissionStatusCompleted,
			isCompleted:  true,
			isFailed:     false,
			isPending:    false,
			isProcessing: false,
		},
		{
			name:         "failed status",
			status:       model.SubmissionStatusFailed,
			isCompleted:  false,
			isFailed:     true,
			isPending:    false,
			isProcessing: false,
		},
		{
			name:         "pending status",
			status:       model.SubmissionStatusPending,
			isCompleted:  false,
			isFailed:     false,
			isPending:    true,
			isProcessing: false,
		},
		{
			name:         "processing status",
			status:       model.SubmissionStatusProcessing,
			isCompleted:  false,
			isFailed:     false,
			isPending:    false,
			isProcessing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submission := &model.FormSubmission{
				FormID: "form123",
				Data:   model.JSON{"test": "data"},
				Status: tt.status,
			}

			require.Equal(t, tt.isCompleted, submission.IsCompleted())
			require.Equal(t, tt.isFailed, submission.IsFailed())
			require.Equal(t, tt.isPending, submission.IsPending())
			require.Equal(t, tt.isProcessing, submission.IsProcessing())
		})
	}
}

func TestFormSubmission_UpdateStatus(t *testing.T) {
	submission := &model.FormSubmission{
		FormID: "form123",
		Data:   model.JSON{"test": "data"},
		Status: model.SubmissionStatusPending,
	}

	// Test status update
	submission.UpdateStatus(model.SubmissionStatusProcessing)
	require.Equal(t, model.SubmissionStatusProcessing, submission.Status)

	// Test status update with metadata
	submission.AddMetadata("status_change", "processing")
	require.Equal(t, "processing", submission.Metadata["status_change"])
}
