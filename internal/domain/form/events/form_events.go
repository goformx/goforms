package form

import (
	"time"

	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// EventType represents the type of form event
type EventType string

const (
	// Form Lifecycle Events
	FormCreatedEventType     EventType = "form.created"
	FormUpdatedEventType     EventType = "form.updated"
	FormDeletedEventType     EventType = "form.deleted"
	FormPublishedEventType   EventType = "form.published"
	FormUnpublishedEventType EventType = "form.unpublished"

	// Form State Events
	FormLoadedEventType EventType = "form.loaded"
	FormReadyEventType  EventType = "form.ready"
	FormDirtyEventType  EventType = "form.dirty"
	FormCleanEventType  EventType = "form.clean"
	FormStateEventType  EventType = "form.state"

	// Form Submission Events
	FormSubmittedEventType EventType = "form.submitted"
	FormValidatedEventType EventType = "form.validated"
	FormProcessedEventType EventType = "form.processed"
	FormErrorEventType     EventType = "form.error"

	// Form Field Events
	FieldFocusedEventType   EventType = "form.field.focused"
	FieldBlurredEventType   EventType = "form.field.blurred"
	FieldChangedEventType   EventType = "form.field.changed"
	FieldValidatedEventType EventType = "form.field.validated"
	FieldEventType          EventType = "form.field"

	// Form Analytics Events
	FormViewedEventType     EventType = "form.viewed"
	FormInteractedEventType EventType = "form.interacted"
	FormAbandonedEventType  EventType = "form.abandoned"
	FormCompletedEventType  EventType = "form.completed"
	AnalyticsEventType      EventType = "form.analytics"
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

// FormStateEvent represents a form state change event
type FormStateEvent struct {
	events.BaseEvent
	FormID    string
	State     string
	ChangedAt time.Time
	Metadata  map[string]interface{}
}

// NewFormStateEvent creates a new form state event
func NewFormStateEvent(formID string, state string, metadata map[string]interface{}) *FormStateEvent {
	return &FormStateEvent{
		BaseEvent: events.NewBaseEvent(string(FormLoadedEventType)),
		FormID:    formID,
		State:     state,
		ChangedAt: time.Now(),
		Metadata:  metadata,
	}
}

// Data returns the event data
func (e *FormStateEvent) Data() any {
	return map[string]any{
		"form_id":    e.FormID,
		"state":      e.State,
		"changed_at": e.ChangedAt,
		"metadata":   e.Metadata,
	}
}

// FieldEvent represents a form field event
type FieldEvent struct {
	events.BaseEvent
	FormID     string
	FieldID    string
	FieldName  string
	FieldValue interface{}
	ChangedAt  time.Time
	Metadata   map[string]interface{}
}

// NewFieldEvent creates a new field event
func NewFieldEvent(formID string, fieldID string, fieldName string, fieldValue interface{}, metadata map[string]interface{}) *FieldEvent {
	return &FieldEvent{
		BaseEvent:  events.NewBaseEvent(string(FieldChangedEventType)),
		FormID:     formID,
		FieldID:    fieldID,
		FieldName:  fieldName,
		FieldValue: fieldValue,
		ChangedAt:  time.Now(),
		Metadata:   metadata,
	}
}

// Data returns the event data
func (e *FieldEvent) Data() any {
	return map[string]any{
		"form_id":     e.FormID,
		"field_id":    e.FieldID,
		"field_name":  e.FieldName,
		"field_value": e.FieldValue,
		"changed_at":  e.ChangedAt,
		"metadata":    e.Metadata,
	}
}

// AnalyticsEvent represents a form analytics event
type AnalyticsEvent struct {
	events.BaseEvent
	FormID    string
	EventType string
	UserID    string
	SessionID string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// NewAnalyticsEvent creates a new analytics event
func NewAnalyticsEvent(formID string, eventType string, userID string, sessionID string, metadata map[string]interface{}) *AnalyticsEvent {
	return &AnalyticsEvent{
		BaseEvent: events.NewBaseEvent(string(FormViewedEventType)),
		FormID:    formID,
		EventType: eventType,
		UserID:    userID,
		SessionID: sessionID,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
}

// Data returns the event data
func (e *AnalyticsEvent) Data() any {
	return map[string]any{
		"form_id":    e.FormID,
		"event_type": e.EventType,
		"user_id":    e.UserID,
		"session_id": e.SessionID,
		"timestamp":  e.Timestamp,
		"metadata":   e.Metadata,
	}
}
