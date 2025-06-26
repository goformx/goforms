// Package form provides form-related domain events and event handling
// functionality for managing form lifecycle and state changes.
package form

import (
	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// EventType represents the type of form event
type EventType string

const (
	// FormCreatedEventType represents a form created event
	FormCreatedEventType EventType = "form.created"
	// FormUpdatedEventType represents a form updated event
	FormUpdatedEventType EventType = "form.updated"
	// FormDeletedEventType represents a form deleted event
	FormDeletedEventType EventType = "form.deleted"
	// FormSubmittedEventType represents a form submitted event
	FormSubmittedEventType EventType = "form.submitted"
	// FormValidatedEventType represents a form validated event
	FormValidatedEventType EventType = "form.validated"
	// FormProcessedEventType represents a form processed event
	FormProcessedEventType EventType = "form.processed"
	// FormErrorEventType represents a form error event
	FormErrorEventType EventType = "form.error"
	// FormStateEventType represents a form state event
	FormStateEventType EventType = "form.state"
	// FieldEventType represents a field event
	FieldEventType EventType = "form.field"
	// AnalyticsEventType represents an analytics event
	AnalyticsEventType EventType = "form.analytics"
)

// Event represents a form-related event
type Event struct {
	events.BaseEvent
	payload any
}

// Ensure Event implements events.Event
var _ events.Event = (*Event)(nil)

// NewEvent creates a new form event
func NewEvent(eventType EventType, payload any) *Event {
	return &Event{
		BaseEvent: events.NewBaseEvent(string(eventType)),
		payload:   payload,
	}
}

// Payload returns the event payload
func (e *Event) Payload() any {
	return e.payload
}

// NewFormCreatedEvent creates a new form created event
func NewFormCreatedEvent(form *model.Form) *Event {
	return NewEvent(FormCreatedEventType, form)
}

// NewFormUpdatedEvent creates a new form updated event
func NewFormUpdatedEvent(form *model.Form) *Event {
	return NewEvent(FormUpdatedEventType, form)
}

// NewFormDeletedEvent creates a new form deleted event
func NewFormDeletedEvent(formID string) *Event {
	return NewEvent(FormDeletedEventType, formID)
}

// NewFormSubmittedEvent creates a new form submitted event
func NewFormSubmittedEvent(submission *model.FormSubmission) *Event {
	return NewEvent(FormSubmittedEventType, submission)
}

// NewFormValidatedEvent creates a new form validated event
func NewFormValidatedEvent(formID string, isValid bool) *Event {
	return NewEvent(FormValidatedEventType, map[string]any{
		"form_id":  formID,
		"is_valid": isValid,
	})
}

// NewFormProcessedEvent creates a new form processed event
func NewFormProcessedEvent(formID, processingID string) *Event {
	return NewEvent(FormProcessedEventType, map[string]string{
		"form_id":       formID,
		"processing_id": processingID,
	})
}

// NewFormErrorEvent creates a new form error event
func NewFormErrorEvent(formID string, err error) *Event {
	return NewEvent(FormErrorEventType, map[string]any{
		"form_id": formID,
		"error":   err.Error(),
	})
}

// NewFormStateEvent creates a new form state event
func NewFormStateEvent(formID, state string) *Event {
	return NewEvent(FormStateEventType, map[string]string{
		"form_id": formID,
		"state":   state,
	})
}

// NewFieldEvent creates a new field event
func NewFieldEvent(formID, fieldID string) *Event {
	return NewEvent(FieldEventType, map[string]string{
		"form_id":  formID,
		"field_id": fieldID,
	})
}

// NewAnalyticsEvent creates a new analytics event
func NewAnalyticsEvent(formID, eventType string) *Event {
	return NewEvent(AnalyticsEventType, map[string]string{
		"form_id":    formID,
		"event_type": eventType,
	})
}
