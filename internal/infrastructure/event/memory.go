// Package event provides in-memory event bus and publisher implementations.
// It implements the domain event interfaces for local event handling.
package event

import (
	"context"
	"fmt"
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
func (b *MemoryEventBus) PublishBatch(ctx context.Context, eventList []events.Event) error {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()

	for _, event := range eventList {
		if err := b.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	return nil
}

// Subscribe subscribes to an event
func (b *MemoryEventBus) Subscribe(
	_ context.Context,
	eventName string,
	handler func(context.Context, events.Event) error,
) error {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()

	if _, exists := b.handlers[eventName]; !exists {
		b.handlers[eventName] = make([]func(context.Context, events.Event) error, 0)
	}

	b.handlers[eventName] = append(b.handlers[eventName], handler)

	return nil
}

// Unsubscribe unsubscribes from an event
func (b *MemoryEventBus) Unsubscribe(_ context.Context, eventName string) error {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()

	delete(b.handlers, eventName)

	return nil
}

// Start starts the event bus
func (b *MemoryEventBus) Start(_ context.Context) error {
	return nil
}

// Stop stops the event bus
func (b *MemoryEventBus) Stop(_ context.Context) error {
	return nil
}

// Health returns the health status of the event bus
func (b *MemoryEventBus) Health(_ context.Context) error {
	return nil
}
