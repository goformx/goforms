//go:generate mockgen -typed -source=event.go -destination=../../../../test/mocks/events/mock_eventbus.go -package=events

package events

import (
	"context"
	"time"
)

// Event represents a domain event
type Event interface {
	// Name returns the event name
	Name() string
	// Timestamp returns when the event occurred
	Timestamp() time.Time
	// Payload returns the event payload
	Payload() any
	// Metadata returns additional event metadata
	Metadata() map[string]any
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	eventName string
	timestamp time.Time
	metadata  map[string]any
}

// NewBaseEvent creates a new base event
func NewBaseEvent(name string) BaseEvent {
	return BaseEvent{
		eventName: name,
		timestamp: time.Now(),
		metadata:  make(map[string]any),
	}
}

// Name returns the event name
func (e BaseEvent) Name() string {
	return e.eventName
}

// Timestamp returns when the event occurred
func (e BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// Metadata returns additional event metadata
func (e BaseEvent) Metadata() map[string]any {
	return e.metadata
}

// Publisher defines the interface for publishing events
type Publisher interface {
	// Publish publishes an event
	Publish(ctx context.Context, event Event) error
	// PublishBatch publishes multiple events
	PublishBatch(ctx context.Context, events []Event) error
}

// Subscriber defines the interface for subscribing to events
type Subscriber interface {
	// Subscribe subscribes to an event
	Subscribe(ctx context.Context, eventName string, handler func(ctx context.Context, event Event) error) error
	// Unsubscribe unsubscribes from an event
	Unsubscribe(ctx context.Context, eventName string) error
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	// Handle handles an event
	Handle(ctx context.Context, event Event) error
}

// EventStore defines the interface for storing events
type EventStore interface {
	// Save saves an event
	Save(ctx context.Context, event Event) error
	// GetByID gets an event by ID
	GetByID(ctx context.Context, id string) (Event, error)
	// GetByType gets events by type
	GetByType(ctx context.Context, eventType string) ([]Event, error)
	// GetByTimeRange gets events within a time range
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]Event, error)
}

// EventBus defines the interface for the event bus
type EventBus interface {
	Publisher
	Subscriber
	// Start starts the event bus
	Start(ctx context.Context) error
	// Stop stops the event bus
	Stop(ctx context.Context) error
	// Health returns the health status of the event bus
	Health(ctx context.Context) error
}
