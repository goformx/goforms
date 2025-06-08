package event

import (
	"context"
	"sync"

	"github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// DefaultMaxEvents is the default maximum number of events to store
	DefaultMaxEvents = 1000
)

// MemoryPublisher is an in-memory implementation of the Publisher interface
type MemoryPublisher struct {
	logger      logging.Logger
	mu          sync.RWMutex
	events      []event.Event
	subscribers []Subscriber
	maxEvents   int
}

// NewMemoryPublisher creates a new in-memory event publisher
func NewMemoryPublisher(logger logging.Logger) *MemoryPublisher {
	return &MemoryPublisher{
		logger:      logger,
		events:      make([]event.Event, 0),
		subscribers: make([]Subscriber, 0),
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
	p.logger.Debug("event published",
		logging.String("name", evt.Name()),
		logging.Time("timestamp", evt.Timestamp()),
	)

	// Notify subscribers
	for _, sub := range p.subscribers {
		go func(s Subscriber) {
			if err := s.Handle(ctx, evt); err != nil {
				p.logger.Error("subscriber error",
					logging.Error(err),
					logging.String("event", evt.Name()),
				)
			}
		}(sub)
	}

	return nil
}

// Subscribe adds a subscriber for events
func (p *MemoryPublisher) Subscribe(subscriber Subscriber) {
	if subscriber == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers = append(p.subscribers, subscriber)
}

// Unsubscribe removes a subscriber
func (p *MemoryPublisher) Unsubscribe(subscriber Subscriber) {
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
