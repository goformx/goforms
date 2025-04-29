package events

import (
	"context"
	"sync"
	"time"
)

// Event represents a domain event
type Event interface {
	// EventID returns the unique identifier of the event
	EventID() string
	// EventType returns the type of the event
	EventType() string
	// Timestamp returns when the event occurred
	Timestamp() time.Time
	// Data returns the event data
	Data() any
}

// EventHandler handles domain events
type EventHandler interface {
	// Handle processes the event
	Handle(ctx context.Context, event Event) error
}

// EventDispatcher dispatches events to registered handlers
type EventDispatcher struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

// RegisterHandler registers an event handler for a specific event type
func (d *EventDispatcher) RegisterHandler(eventType string, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Dispatch dispatches an event to all registered handlers
func (d *EventDispatcher) Dispatch(ctx context.Context, event Event) error {
	d.mu.RLock()
	handlers := d.handlers[event.EventType()]
	d.mu.RUnlock()

	var errs []error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return &DispatchError{Errors: errs}
	}

	return nil
}

// DispatchError represents errors that occurred during event dispatch
type DispatchError struct {
	Errors []error
}

func (e *DispatchError) Error() string {
	return "one or more errors occurred during event dispatch"
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	eventID   string
	eventType string
	timestamp time.Time
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType string) BaseEvent {
	return BaseEvent{
		eventID:   generateEventID(),
		eventType: eventType,
		timestamp: time.Now(),
	}
}

// EventID returns the event ID
func (e BaseEvent) EventID() string {
	return e.eventID
}

// EventType returns the event type
func (e BaseEvent) EventType() string {
	return e.eventType
}

// Timestamp returns the event timestamp
func (e BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405.000000000")
}

// EventStore stores events for later retrieval
type EventStore interface {
	// Save saves an event
	Save(ctx context.Context, event Event) error
	// Load loads events for an aggregate
	Load(ctx context.Context, aggregateID string) ([]Event, error)
}

// InMemoryEventStore implements EventStore using memory
type InMemoryEventStore struct {
	events map[string][]Event
	mu     sync.RWMutex
}

// NewInMemoryEventStore creates a new in-memory event store
func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: make(map[string][]Event),
	}
}

// Save saves an event
func (s *InMemoryEventStore) Save(ctx context.Context, event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// For simplicity, we use the event ID as the aggregate ID
	// In a real application, you would have a proper aggregate ID
	s.events[event.EventID()] = append(s.events[event.EventID()], event)
	return nil
}

// Load loads events for an aggregate
func (s *InMemoryEventStore) Load(ctx context.Context, aggregateID string) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events, exists := s.events[aggregateID]
	if !exists {
		return nil, nil
	}

	return events, nil
} 