package event

import (
	"context"
	"errors"
	"time"

	"github.com/goformx/goforms/internal/domain/form/event"
)

var (
	// ErrInvalidEvent is returned when an event is invalid
	ErrInvalidEvent = errors.New("invalid event")
	// ErrEventValidation is returned when event validation fails
	ErrEventValidation = errors.New("event validation failed")
)

// Publisher defines the interface for publishing events
type Publisher interface {
	// Publish publishes an event
	Publish(ctx context.Context, event event.Event) error
	// Subscribe adds a subscriber for events
	Subscribe(subscriber Subscriber)
	// Unsubscribe removes a subscriber
	Unsubscribe(subscriber Subscriber)
}

// Subscriber defines the interface for event subscribers
type Subscriber interface {
	// Handle handles an event
	Handle(ctx context.Context, event event.Event) error
}

// Event represents a domain event
type Event interface {
	// Name returns the event name
	Name() string
	// Timestamp returns when the event occurred
	Timestamp() time.Time
	// Payload returns the event payload
	Payload() any
}
