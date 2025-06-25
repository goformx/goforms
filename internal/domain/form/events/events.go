// Package form provides form-related domain events and event handling
// functionality for managing form lifecycle and state changes.
package form

import (
	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// FormEventType represents the type of form event
type FormEventType string

const (
	FormCreatedEventType   FormEventType = "form.created"
	FormUpdatedEventType   FormEventType = "form.updated"
	FormDeletedEventType   FormEventType = "form.deleted"
	FormSubmittedEventType FormEventType = "form.submitted"
	FormValidatedEventType FormEventType = "form.validated"
	FormProcessedEventType FormEventType = "form.processed"
	FormErrorEventType     FormEventType = "form.error"
	FormStateEventType     FormEventType = "form.state"
	FieldEventType         FormEventType = "form.field"
	AnalyticsEventType     FormEventType = "form.analytics"
)

// FormEvent represents a form-related event
type FormEvent struct {
	events.BaseEvent
	payload any
}

// Ensure FormEvent implements events.Event
var _ events.Event = (*FormEvent)(nil)

// NewFormEvent creates a new form event
func NewFormEvent(eventType FormEventType, payload any) *FormEvent {
	return &FormEvent{
		BaseEvent: events.NewBaseEvent(string(eventType)),
		payload:   payload,
	}
}

// Payload returns the event payload
func (e *FormEvent) Payload() any {
	return e.payload
}

// NewFormCreatedEvent creates a new form created event
func NewFormCreatedEvent(form *model.Form) *FormEvent {
	return NewFormEvent(FormCreatedEventType, form)
}

// NewFormUpdatedEvent creates a new form updated event
func NewFormUpdatedEvent(form *model.Form) *FormEvent {
	return NewFormEvent(FormUpdatedEventType, form)
}

// NewFormDeletedEvent creates a new form deleted event
func NewFormDeletedEvent(formID string) *FormEvent {
	return NewFormEvent(FormDeletedEventType, formID)
}

// NewFormSubmittedEvent creates a new form submitted event
func NewFormSubmittedEvent(submission *model.FormSubmission) *FormEvent {
	return NewFormEvent(FormSubmittedEventType, submission)
}

// NewFormValidatedEvent creates a new form validated event
func NewFormValidatedEvent(formID string, isValid bool) *FormEvent {
	return NewFormEvent(FormValidatedEventType, map[string]any{
		"form_id":  formID,
		"is_valid": isValid,
	})
}

// NewFormProcessedEvent creates a new form processed event
func NewFormProcessedEvent(formID, processingID string) *FormEvent {
	return NewFormEvent(FormProcessedEventType, map[string]string{
		"form_id":       formID,
		"processing_id": processingID,
	})
}

// NewFormErrorEvent creates a new form error event
func NewFormErrorEvent(formID string, err error) *FormEvent {
	return NewFormEvent(FormErrorEventType, map[string]any{
		"form_id": formID,
		"error":   err.Error(),
	})
}

// NewFormStateEvent creates a new form state event
func NewFormStateEvent(formID, state string) *FormEvent {
	return NewFormEvent(FormStateEventType, map[string]string{
		"form_id": formID,
		"state":   state,
	})
}

// NewFieldEvent creates a new field event
func NewFieldEvent(formID, fieldID string) *FormEvent {
	return NewFormEvent(FieldEventType, map[string]string{
		"form_id":  formID,
		"field_id": fieldID,
	})
}

// NewAnalyticsEvent creates a new analytics event
func NewAnalyticsEvent(formID, eventType string) *FormEvent {
	return NewFormEvent(AnalyticsEventType, map[string]string{
		"form_id":    formID,
		"event_type": eventType,
	})
}
