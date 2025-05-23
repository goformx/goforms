package form

import (
	"time"

	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// EventType represents the type of form event
type EventType string

const (
	// FormSubmittedEventType is emitted when a form is submitted
	FormSubmittedEventType EventType = "form.submitted"
	// FormValidatedEventType is emitted when a form is validated
	FormValidatedEventType EventType = "form.validated"
	// FormProcessedEventType is emitted when a form is processed
	FormProcessedEventType EventType = "form.processed"
	// FormErrorEventType is emitted when an error occurs during form processing
	FormErrorEventType EventType = "form.error"
)

// SubmittedEvent represents a form submission event
type SubmittedEvent struct {
	events.BaseEvent
	FormID      string
	Submission  *model.FormSubmission
	SubmittedAt time.Time
}

// NewFormSubmittedEvent creates a new form submitted event
func NewFormSubmittedEvent(formID string, submission *model.FormSubmission) *SubmittedEvent {
	return &SubmittedEvent{
		BaseEvent:   events.NewBaseEvent(string(FormSubmittedEventType)),
		FormID:      formID,
		Submission:  submission,
		SubmittedAt: time.Now(),
	}
}

// Data returns the event data
func (e *SubmittedEvent) Data() any {
	return e.Submission
}

// ValidatedEvent represents a form validation event
type ValidatedEvent struct {
	events.BaseEvent
	FormID      string
	Submission  *model.FormSubmission
	ValidatedAt time.Time
	IsValid     bool
	Errors      []error
}

// NewFormValidatedEvent creates a new form validated event
func NewFormValidatedEvent(
	formID string,
	submission *model.FormSubmission,
	isValid bool,
	errors []error,
) *ValidatedEvent {
	return &ValidatedEvent{
		BaseEvent:   events.NewBaseEvent(string(FormValidatedEventType)),
		FormID:      formID,
		Submission:  submission,
		ValidatedAt: time.Now(),
		IsValid:     isValid,
		Errors:      errors,
	}
}

// Data returns the event data
func (e *ValidatedEvent) Data() any {
	return map[string]any{
		"form_id":      e.FormID,
		"submission":   e.Submission,
		"is_valid":     e.IsValid,
		"errors":       e.Errors,
		"validated_at": e.ValidatedAt,
	}
}

// ProcessedEvent represents a form processing event
type ProcessedEvent struct {
	events.BaseEvent
	FormID       string
	Submission   *model.FormSubmission
	ProcessedAt  time.Time
	ProcessingID string
}

// NewFormProcessedEvent creates a new form processed event
func NewFormProcessedEvent(formID string, submission *model.FormSubmission, processingID string) *ProcessedEvent {
	return &ProcessedEvent{
		BaseEvent:    events.NewBaseEvent(string(FormProcessedEventType)),
		FormID:       formID,
		Submission:   submission,
		ProcessedAt:  time.Now(),
		ProcessingID: processingID,
	}
}

// Data returns the event data
func (e *ProcessedEvent) Data() any {
	return map[string]any{
		"form_id":       e.FormID,
		"submission":    e.Submission,
		"processed_at":  e.ProcessedAt,
		"processing_id": e.ProcessingID,
	}
}

// ErrorEvent represents a form error event
type ErrorEvent struct {
	events.BaseEvent
	FormID     string
	Submission *model.FormSubmission
	Error      error
	OccurredAt time.Time
	ErrorType  string
}

// NewFormErrorEvent creates a new form error event
func NewFormErrorEvent(formID string, submission *model.FormSubmission, err error, errorType string) *ErrorEvent {
	return &ErrorEvent{
		BaseEvent:  events.NewBaseEvent(string(FormErrorEventType)),
		FormID:     formID,
		Submission: submission,
		Error:      err,
		OccurredAt: time.Now(),
		ErrorType:  errorType,
	}
}

// Data returns the event data
func (e *ErrorEvent) Data() any {
	return map[string]any{
		"form_id":     e.FormID,
		"submission":  e.Submission,
		"error":       e.Error,
		"occurred_at": e.OccurredAt,
		"error_type":  e.ErrorType,
	}
}
