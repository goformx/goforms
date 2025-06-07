package model

import (
	"time"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/validation"
)

// FormSubmission represents a form submission
type FormSubmission struct {
	ID          string           `json:"id"`
	FormID      string           `json:"form_id"`
	Data        JSON             `json:"data"`
	SubmittedAt time.Time        `json:"submitted_at"`
	Status      SubmissionStatus `json:"status"`
	Metadata    JSON             `json:"metadata"`
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

// Validate validates the form submission
func (s *FormSubmission) Validate() error {
	validator := validation.New()
	if err := validator.Struct(s); err != nil {
		return errors.Wrap(err, errors.ErrCodeValidation, "form submission validation failed")
	}

	if s.Data == nil {
		return errors.New(errors.ErrCodeValidation, "form data is required", nil)
	}

	if len(s.Data) == 0 {
		return errors.New(errors.ErrCodeValidation, "form data cannot be empty", nil)
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
		s.Metadata = make(JSON)
	}
	s.Metadata[key] = value
}

// GetMetadata returns the metadata value for a key
func (s *FormSubmission) GetMetadata(key string) string {
	if s.Metadata == nil {
		return ""
	}
	if val, ok := s.Metadata[key].(string); ok {
		return val
	}
	return ""
}
