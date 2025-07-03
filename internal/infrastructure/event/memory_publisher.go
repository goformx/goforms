// Package event provides in-memory event bus and publisher implementations.
// It implements the domain event interfaces for local event handling.
package event

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/goformx/goforms/internal/domain/common/events"
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
	events    []events.Event
	handlers  map[string][]func(ctx context.Context, event events.Event) error
	maxEvents int
}

// NewMemoryPublisher creates a new in-memory event publisher
func NewMemoryPublisher(logger logging.Logger) *MemoryPublisher {
	return &MemoryPublisher{
		logger:    logger,
		events:    make([]events.Event, 0),
		handlers:  make(map[string][]func(ctx context.Context, event events.Event) error),
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
	evt events.Event,
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
			go func(h func(ctx context.Context, event events.Event) error) {
				if err := h(ctx, evt); err != nil {
					p.logger.Error("failed to handle event", "error", err, "event", evt.Name())
				}
			}(handler)
		}
	}

	return nil
}

// PublishBatch publishes multiple events to memory
func (p *MemoryPublisher) PublishBatch(
	ctx context.Context,
	evts []events.Event,
) error {
	if evts == nil || len(evts) == 0 {
		return nil
	}

	for _, evt := range evts {
		if err := p.Publish(ctx, evt); err != nil {
			return err
		}
	}

	return nil
}

// Subscribe adds a handler for a specific event type
func (p *MemoryPublisher) Subscribe(
	_ context.Context,
	eventName string,
	handler func(ctx context.Context, event events.Event) error,
) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.handlers[eventName]; !ok {
		p.handlers[eventName] = make([]func(ctx context.Context, event events.Event) error, 0)
	}

	p.handlers[eventName] = append(p.handlers[eventName], handler)

	return nil
}

// GetEvents returns all published events
func (p *MemoryPublisher) GetEvents() []events.Event {
	p.mu.RLock()
	defer p.mu.RUnlock()

	events := make([]events.Event, len(p.events))
	copy(events, p.events)

	return events
}

// ClearEvents clears all published events
func (p *MemoryPublisher) ClearEvents() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.events = make([]events.Event, 0)
}

// Unsubscribe removes all handlers for a specific event type (no-op for in-memory)
func (p *MemoryPublisher) Unsubscribe(ctx context.Context, eventName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.handlers, eventName)

	return nil
}

// Start starts the event bus (no-op for in-memory)
func (p *MemoryPublisher) Start(ctx context.Context) error {
	return nil
}

// Stop stops the event bus (no-op for in-memory)
func (p *MemoryPublisher) Stop(ctx context.Context) error {
	return nil
}

// Health returns the health status of the event bus (always healthy for in-memory)
func (p *MemoryPublisher) Health(ctx context.Context) error {
	return nil
}
