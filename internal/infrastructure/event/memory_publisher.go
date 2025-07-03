// Package event provides in-memory event bus and publisher implementations.
// It implements the domain event interfaces for local event handling.
package event

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// DefaultMaxEvents is the default maximum number of events to store
	DefaultMaxEvents = 1000
)

// ErrInvalidEvent is returned when an invalid event is published
var ErrInvalidEvent = errors.New("invalid event")

// MemoryPublisher is an in-memory implementation of the events.Publisher interface
type MemoryPublisher struct {
	logger    logging.Logger
	mu        sync.RWMutex
	events    []event.Event
	handlers  map[string][]func(ctx context.Context, event event.Event) error
	maxEvents int
}

// NewMemoryPublisher creates a new in-memory event publisher
func NewMemoryPublisher(logger logging.Logger) event.Publisher {
	return &MemoryPublisher{
		logger:    logger,
		events:    make([]event.Event, 0),
		handlers:  make(map[string][]func(ctx context.Context, event event.Event) error),
		maxEvents: DefaultMaxEvents,
	}
}

// WithMaxEvents sets the maximum number of events to store
func (p *MemoryPublisher) WithMaxEvents(maxEvents int) *MemoryPublisher {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.maxEvents = maxEvents

	return p
}

// Publish publishes an event to memory
func (p *MemoryPublisher) Publish(
	ctx context.Context,
	evt event.Event,
) error {
	if evt == nil {
		return ErrInvalidEvent
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we need to trim old events
	if len(p.events) >= p.maxEvents {
		p.events = p.events[1:]
	}

	p.events = append(p.events, evt)
	p.logger.Debug("publishing event", "name", evt.Name(), "time", time.Now())

	// Notify handlers
	if handlers, ok := p.handlers[evt.Name()]; ok {
		for _, handler := range handlers {
			go func(h func(ctx context.Context, event event.Event) error) {
				if err := h(ctx, evt); err != nil {
					p.logger.Error("failed to handle event", "error", err, "event", evt.Name())
				}
			}(handler)
		}
	}

	return nil
}

// Subscribe adds a handler for a specific event type
func (p *MemoryPublisher) Subscribe(
	_ context.Context,
	eventName string,
	handler func(ctx context.Context, event event.Event) error,
) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.handlers[eventName]; !ok {
		p.handlers[eventName] = make([]func(ctx context.Context, event event.Event) error, 0)
	}

	p.handlers[eventName] = append(p.handlers[eventName], handler)

	return nil
}

// GetEvents returns all published events
func (p *MemoryPublisher) GetEvents() []event.Event {
	p.mu.RLock()
	defer p.mu.RUnlock()

	events := make([]event.Event, len(p.events))
	copy(events, p.events)

	return events
}

// ClearEvents clears all published events
func (p *MemoryPublisher) ClearEvents() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.events = make([]event.Event, 0)
}
