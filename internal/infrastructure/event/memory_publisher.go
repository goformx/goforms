package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// MemoryPublisher is an in-memory implementation of the Publisher interface
type MemoryPublisher struct {
	logger logging.Logger
	mu     sync.RWMutex
	events []event.Event
}

// NewMemoryPublisher creates a new in-memory event publisher
func NewMemoryPublisher(logger logging.Logger) *MemoryPublisher {
	return &MemoryPublisher{
		logger: logger,
		events: make([]event.Event, 0),
	}
}

// Publish publishes an event to memory
func (p *MemoryPublisher) Publish(ctx context.Context, evt event.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.events = append(p.events, evt)
	p.logger.Debug(fmt.Sprintf("event published: %s at %s", evt.Name(), evt.Timestamp()))
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
