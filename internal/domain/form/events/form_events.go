package form

import (
	"time"

	"github.com/jonesrussell/goforms/internal/domain/common/events"
	"github.com/jonesrussell/goforms/internal/domain/form/model"
)

// FormEventType represents the type of form-related events
type FormEventType string

const (
	// FormSubmittedEventType is emitted when a form is submitted
	FormSubmittedEventType FormEventType = "form.submitted"
	// FormValidatedEventType is emitted when a form is validated
	FormValidatedEventType FormEventType = "form.validated"
	// FormProcessedEventType is emitted when a form is processed
	FormProcessedEventType FormEventType = "form.processed"
	// FormErrorEventType is emitted when an error occurs during form processing
	FormErrorEventType FormEventType = "form.error"
)

// FormSubmittedEvent represents a form submission event
type FormSubmittedEvent struct {
	events.BaseEvent
	FormID      string
	Submission  *model.FormSubmission
	SubmittedAt time.Time
}

// NewFormSubmittedEvent creates a new form submitted event
func NewFormSubmittedEvent(formID string, submission *model.FormSubmission) *FormSubmittedEvent {
	return &FormSubmittedEvent{
		BaseEvent:   events.NewBaseEvent(string(FormSubmittedEventType)),
		FormID:      formID,
		Submission:  submission,
		SubmittedAt: time.Now(),
	}
}

// Data returns the event data
func (e *FormSubmittedEvent) Data() any {
	return e.Submission
}

// FormValidatedEvent represents a form validation event
type FormValidatedEvent struct {
	events.BaseEvent
	FormID      string
	Submission  *model.FormSubmission
	ValidatedAt time.Time
	IsValid     bool
	Errors      []error
}

// NewFormValidatedEvent creates a new form validated event
func NewFormValidatedEvent(formID string, submission *model.FormSubmission, isValid bool, errors []error) *FormValidatedEvent {
	return &FormValidatedEvent{
		BaseEvent:   events.NewBaseEvent(string(FormValidatedEventType)),
		FormID:      formID,
		Submission:  submission,
		ValidatedAt: time.Now(),
		IsValid:     isValid,
		Errors:      errors,
	}
}

// Data returns the event data
func (e *FormValidatedEvent) Data() any {
	return map[string]any{
		"form_id":      e.FormID,
		"submission":   e.Submission,
		"is_valid":     e.IsValid,
		"errors":       e.Errors,
		"validated_at": e.ValidatedAt,
	}
}

// FormProcessedEvent represents a form processing event
type FormProcessedEvent struct {
	events.BaseEvent
	FormID       string
	Submission   *model.FormSubmission
	ProcessedAt  time.Time
	ProcessingID string
}

// NewFormProcessedEvent creates a new form processed event
func NewFormProcessedEvent(formID string, submission *model.FormSubmission, processingID string) *FormProcessedEvent {
	return &FormProcessedEvent{
		BaseEvent:    events.NewBaseEvent(string(FormProcessedEventType)),
		FormID:       formID,
		Submission:   submission,
		ProcessedAt:  time.Now(),
		ProcessingID: processingID,
	}
}

// Data returns the event data
func (e *FormProcessedEvent) Data() any {
	return map[string]any{
		"form_id":       e.FormID,
		"submission":    e.Submission,
		"processed_at":  e.ProcessedAt,
		"processing_id": e.ProcessingID,
	}
}

// FormErrorEvent represents a form error event
type FormErrorEvent struct {
	events.BaseEvent
	FormID      string
	Submission  *model.FormSubmission
	Error       error
	OccurredAt  time.Time
	ErrorType   string
}

// NewFormErrorEvent creates a new form error event
func NewFormErrorEvent(formID string, submission *model.FormSubmission, err error, errorType string) *FormErrorEvent {
	return &FormErrorEvent{
		BaseEvent:   events.NewBaseEvent(string(FormErrorEventType)),
		FormID:      formID,
		Submission:  submission,
		Error:       err,
		OccurredAt:  time.Now(),
		ErrorType:   errorType,
	}
}

// Data returns the event data
func (e *FormErrorEvent) Data() any {
	return map[string]any{
		"form_id":     e.FormID,
		"submission":  e.Submission,
		"error":       e.Error,
		"occurred_at": e.OccurredAt,
		"error_type":  e.ErrorType,
	}
} 