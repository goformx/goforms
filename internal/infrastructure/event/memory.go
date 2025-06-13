package event

import (
	"context"
	"sync"

	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// MemoryEventBus implements events.EventBus using in-memory storage
type MemoryEventBus struct {
	logger     logging.Logger
	handlers   map[string][]func(context.Context, events.Event) error
	handlersMu sync.RWMutex
}

// NewMemoryEventBus creates a new memory-based event bus
func NewMemoryEventBus(logger logging.Logger) events.EventBus {
	return &MemoryEventBus{
		logger:   logger,
		handlers: make(map[string][]func(context.Context, events.Event) error),
	}
}

// Publish publishes an event to all subscribers
func (b *MemoryEventBus) Publish(ctx context.Context, event events.Event) error {
	b.handlersMu.RLock()
	handlers := b.handlers[event.Name()]
	b.handlersMu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			b.logger.Error("failed to handle event",
				"event", event.Name(),
				"error", err,
			)
		}
	}
	return nil
}

// PublishBatch publishes multiple events
func (b *MemoryEventBus) PublishBatch(ctx context.Context, events []events.Event) error {
	for _, event := range events {
		if err := b.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe subscribes to an event
func (b *MemoryEventBus) Subscribe(ctx context.Context, eventName string, handler func(context.Context, events.Event) error) error {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()

	b.handlers[eventName] = append(b.handlers[eventName], handler)
	return nil
}

// Unsubscribe unsubscribes from an event
func (b *MemoryEventBus) Unsubscribe(ctx context.Context, eventName string) error {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()

	delete(b.handlers, eventName)
	return nil
}

// Start starts the event bus
func (b *MemoryEventBus) Start(ctx context.Context) error {
	return nil
}

// Stop stops the event bus
func (b *MemoryEventBus) Stop(ctx context.Context) error {
	return nil
}

// Health returns the health status of the event bus
func (b *MemoryEventBus) Health(ctx context.Context) error {
	return nil
}
