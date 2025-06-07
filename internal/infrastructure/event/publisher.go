package event

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/domain/form/event"
)

// Publisher defines the interface for publishing events
type Publisher interface {
	// Publish publishes an event
	Publish(ctx context.Context, event event.Event) error
}

// Event represents a domain event
type Event interface {
	// Name returns the event name
	Name() string
	// Timestamp returns when the event occurred
	Timestamp() time.Time
	// Payload returns the event payload
	Payload() interface{}
}
