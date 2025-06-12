package event

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/domain/form/model"
)

// Event represents a domain event
type Event interface {
	// Name returns the event name
	Name() string
	// Timestamp returns when the event occurred
	Timestamp() time.Time
	// Payload returns the event payload
	Payload() any
}

// Publisher defines the interface for publishing events
type Publisher interface {
	// Publish publishes an event
	Publish(ctx context.Context, event Event) error
}

// Subscriber defines the interface for subscribing to events
type Subscriber interface {
	// Subscribe subscribes to an event
	Subscribe(ctx context.Context, eventName string, handler func(ctx context.Context, event Event) error) error
}

// FormCreatedEvent represents a form creation event
type FormCreatedEvent struct {
	Form      *model.Form
	timestamp time.Time
}

// NewFormCreatedEvent creates a new form created event
func NewFormCreatedEvent(form *model.Form) Event {
	return &FormCreatedEvent{
		Form:      form,
		timestamp: time.Now(),
	}
}

func (e *FormCreatedEvent) Name() string {
	return "form.created"
}

func (e *FormCreatedEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e *FormCreatedEvent) Payload() any {
	return e.Form
}

// FormUpdatedEvent represents a form update event
type FormUpdatedEvent struct {
	Form      *model.Form
	timestamp time.Time
}

// NewFormUpdatedEvent creates a new form updated event
func NewFormUpdatedEvent(form *model.Form) Event {
	return &FormUpdatedEvent{
		Form:      form,
		timestamp: time.Now(),
	}
}

func (e *FormUpdatedEvent) Name() string {
	return "form.updated"
}

func (e *FormUpdatedEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e *FormUpdatedEvent) Payload() any {
	return e.Form
}

// FormDeletedEvent represents a form deletion event
type FormDeletedEvent struct {
	FormID    string
	timestamp time.Time
}

// NewFormDeletedEvent creates a new form deleted event
func NewFormDeletedEvent(formID string) Event {
	return &FormDeletedEvent{
		FormID:    formID,
		timestamp: time.Now(),
	}
}

func (e *FormDeletedEvent) Name() string {
	return "form.deleted"
}

func (e *FormDeletedEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e *FormDeletedEvent) Payload() any {
	return e.FormID
}

// FormSubmissionCreatedEvent represents a form submission creation event
type FormSubmissionCreatedEvent struct {
	Submission *model.FormSubmission
	timestamp  time.Time
}

// NewFormSubmissionCreatedEvent creates a new form submission created event
func NewFormSubmissionCreatedEvent(submission *model.FormSubmission) Event {
	return &FormSubmissionCreatedEvent{
		Submission: submission,
		timestamp:  time.Now(),
	}
}

func (e *FormSubmissionCreatedEvent) Name() string {
	return "form.submission.created"
}

func (e *FormSubmissionCreatedEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e *FormSubmissionCreatedEvent) Payload() any {
	return e.Submission
}
