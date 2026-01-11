// Package model contains domain models and error definitions for forms.
package model

import (
	"time"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// FormSubmission represents a form submission
type FormSubmission struct {
	ID          string           `gorm:"column:uuid;primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	FormID      string           `gorm:"not null;index;type:uuid"                                   json:"form_id"`
	Data        JSON             `gorm:"type:jsonb;not null"                                        json:"data"`
	SubmittedAt time.Time        `gorm:"not null"                                                   json:"submitted_at"`
	Status      SubmissionStatus `gorm:"not null;size:20"                                           json:"status"`
	Metadata    JSON             `gorm:"type:jsonb"                                                 json:"metadata"`
	CreatedAt   time.Time        `gorm:"not null;autoCreateTime"                                    json:"created_at"`
	UpdatedAt   time.Time        `gorm:"not null;autoUpdateTime"                                    json:"updated_at"`
}

// GetID returns the submission's ID
func (fs *FormSubmission) GetID() string {
	return fs.ID
}

// SetID sets the submission's ID
func (fs *FormSubmission) SetID(id string) {
	fs.ID = id
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
func (fs *FormSubmission) Validate() error {
	if fs.FormID == "" {
		return errors.New(errors.ErrCodeValidation, "form ID is required", nil)
	}

	if fs.Data == nil {
		return errors.New(errors.ErrCodeValidation, "submission data is required", nil)
	}

	if fs.Status == "" {
		fs.Status = SubmissionStatusPending
	}

	return nil
}

// Sanitize sanitizes the form submission data using the provided sanitizer
func (fs *FormSubmission) Sanitize(sanitizer sanitization.ServiceInterface) {
	if fs.Data != nil {
		for key, value := range fs.Data {
			if strValue, ok := value.(string); ok {
				fs.Data[key] = sanitizer.String(strValue)
			}
		}
	}

	if fs.Metadata != nil {
		for key, value := range fs.Metadata {
			if strValue, ok := value.(string); ok {
				fs.Metadata[key] = sanitizer.String(strValue)
			}
		}
	}
}

// SetStatus sets the submission status
func (fs *FormSubmission) SetStatus(status SubmissionStatus) {
	fs.Status = status
	fs.UpdatedAt = time.Now()
}

// IsCompleted returns whether the submission is completed
func (fs *FormSubmission) IsCompleted() bool {
	return fs.Status == SubmissionStatusCompleted
}

// IsFailed returns whether the submission failed
func (fs *FormSubmission) IsFailed() bool {
	return fs.Status == SubmissionStatusFailed
}

// IsPending returns whether the submission is pending
func (fs *FormSubmission) IsPending() bool {
	return fs.Status == SubmissionStatusPending
}

// IsProcessing returns whether the submission is being processed
func (fs *FormSubmission) IsProcessing() bool {
	return fs.Status == SubmissionStatusProcessing
}

// UpdateStatus updates the submission status
func (fs *FormSubmission) UpdateStatus(status SubmissionStatus) {
	fs.Status = status
}

// AddMetadata adds metadata to the submission
func (fs *FormSubmission) AddMetadata(key, value string) {
	if fs.Metadata == nil {
		fs.Metadata = make(JSON)
	}

	fs.Metadata[key] = value
}

// GetMetadata returns the metadata value for a key
func (fs *FormSubmission) GetMetadata(key string) string {
	if fs.Metadata == nil {
		return ""
	}

	if val, ok := fs.Metadata[key].(string); ok {
		return val
	}

	return ""
}
