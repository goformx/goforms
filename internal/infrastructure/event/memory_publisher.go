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
	logger      logging.Logger
	mu          sync.RWMutex
	events      []event.Event
	subscribers []event.Subscriber
	maxEvents   int
}

// NewMemoryPublisher creates a new in-memory event publisher
func NewMemoryPublisher(logger logging.Logger) event.Publisher {
	return &MemoryPublisher{
		logger:      logger,
		events:      make([]event.Event, 0),
		subscribers: make([]event.Subscriber, 0),
		maxEvents:   DefaultMaxEvents,
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
func (p *MemoryPublisher) Publish(ctx context.Context, evt event.Event) error {
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

	// TODO(goforms): Refactor subscriber notification to match new event.Subscriber interface (see TODO.md 'To Discuss')
	// for _, sub := range p.subscribers {
	// 	go func(s event.Subscriber) {
	// 		if err := s.Handle(ctx, evt); err != nil {
	// 			p.logger.Error("failed to publish event", "error", err, "event", evt.Name())
	// 		}
	// 	}(sub)
	// }

	return nil
}

// Subscribe adds a subscriber for events
func (p *MemoryPublisher) Subscribe(subscriber event.Subscriber) {
	if subscriber == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers = append(p.subscribers, subscriber)
}

// Unsubscribe removes a subscriber
func (p *MemoryPublisher) Unsubscribe(subscriber event.Subscriber) {
	if subscriber == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, sub := range p.subscribers {
		if sub == subscriber {
			p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
			break
		}
	}
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
