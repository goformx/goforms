package model

import (
	"time"

	"github.com/jonesrussell/goforms/internal/domain/common/errors"
	"github.com/jonesrussell/goforms/internal/domain/common/validation"
)

// FormSubmission represents a form submission
type FormSubmission struct {
	ID          string            `json:"id"`
	FormID      uint              `json:"form_id"`
	Data        map[string]any    `json:"data"`
	SubmittedAt time.Time         `json:"submitted_at"`
	Status      SubmissionStatus  `json:"status"`
	Metadata    map[string]string `json:"metadata"`
}

// SubmissionStatus represents the status of a form submission
type SubmissionStatus string

const (
	// SubmissionStatusPending indicates the submission is pending processing
	SubmissionStatusPending SubmissionStatus = "pending"
	// SubmissionStatusProcessing indicates the submission is being processed
	SubmissionStatusProcessing SubmissionStatus = "processing"
	// SubmissionStatusCompleted indicates the submission has been processed successfully
	SubmissionStatusCompleted SubmissionStatus = "completed"
	// SubmissionStatusFailed indicates the submission processing failed
	SubmissionStatusFailed SubmissionStatus = "failed"
)

// NewFormSubmission creates a new form submission
func NewFormSubmission(formID uint, data map[string]any, metadata map[string]string) (*FormSubmission, error) {
	submission := &FormSubmission{
		ID:          generateSubmissionID(),
		FormID:      formID,
		Data:        data,
		SubmittedAt: time.Now(),
		Status:      SubmissionStatusPending,
		Metadata:    metadata,
	}

	if err := submission.Validate(); err != nil {
		return nil, err
	}

	return submission, nil
}

// Validate validates the form submission
func (s *FormSubmission) Validate() error {
	validator := validation.New()
	if err := validator.ValidateStruct(s); err != nil {
		return errors.Wrap(err, errors.ErrCodeValidation, "form submission validation failed")
	}

	if s.Data == nil {
		return errors.New(errors.ErrCodeValidation, "form data is required")
	}

	if len(s.Data) == 0 {
		return errors.New(errors.ErrCodeValidation, "form data cannot be empty")
	}

	return nil
}

// UpdateStatus updates the submission status
func (s *FormSubmission) UpdateStatus(status SubmissionStatus) {
	s.Status = status
}

// AddMetadata adds metadata to the submission
func (s *FormSubmission) AddMetadata(key, value string) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}
	s.Metadata[key] = value
}

// GetMetadata returns the metadata value for a key
func (s *FormSubmission) GetMetadata(key string) string {
	if s.Metadata == nil {
		return ""
	}
	return s.Metadata[key]
}

// generateSubmissionID generates a unique submission ID
func generateSubmissionID() string {
	return time.Now().Format("20060102150405.000000000")
}
