package events

import (
	"context"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// HandlerConfig contains configuration for event handlers
type HandlerConfig struct {
	Logger     logging.Logger
	RetryCount int
	Timeout    time.Duration
}

// BaseHandler provides common functionality for event handlers
type BaseHandler struct {
	config HandlerConfig
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(config HandlerConfig) *BaseHandler {
	if config.RetryCount <= 0 {
		config.RetryCount = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &BaseHandler{
		config: config,
	}
}

// HandleWithRetry handles an event with retry logic
func (h *BaseHandler) HandleWithRetry(ctx context.Context, event Event, handler func(ctx context.Context, event Event) error) error {
	var lastErr error
	for i := 0; i < h.config.RetryCount; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := handler(ctx, event); err != nil {
				lastErr = err
				h.config.Logger.Warn("event handling failed, retrying",
					"event", event.Name(),
					"attempt", i+1,
					"error", err,
				)
				continue
			}
			return nil
		}
	}
	return fmt.Errorf("event handling failed after %d attempts: %v", h.config.RetryCount, lastErr)
}

// LogEvent logs an event
func (h *BaseHandler) LogEvent(event Event, level string, message string, fields ...any) {
	logger := h.config.Logger.With(
		"event", event.Name(),
		"timestamp", event.Timestamp(),
		"metadata", event.Metadata(),
	)

	switch level {
	case "debug":
		logger.Debug(message, fields...)
	case "info":
		logger.Info(message, fields...)
	case "warn":
		logger.Warn(message, fields...)
	case "error":
		logger.Error(message, fields...)
	default:
		logger.Info(message, fields...)
	}
}

// EventHandlerRegistry manages event handlers
type EventHandlerRegistry struct {
	handlers map[string][]EventHandler
}

// NewEventHandlerRegistry creates a new event handler registry
func NewEventHandlerRegistry() *EventHandlerRegistry {
	return &EventHandlerRegistry{
		handlers: make(map[string][]EventHandler),
	}
}

// RegisterHandler registers an event handler
func (r *EventHandlerRegistry) RegisterHandler(eventName string, handler EventHandler) {
	r.handlers[eventName] = append(r.handlers[eventName], handler)
}

// GetHandlers gets handlers for an event
func (r *EventHandlerRegistry) GetHandlers(eventName string) []EventHandler {
	return r.handlers[eventName]
}

// HandleEvent handles an event by calling all registered handlers
func (r *EventHandlerRegistry) HandleEvent(ctx context.Context, event Event) error {
	handlers := r.GetHandlers(event.Name())
	if len(handlers) == 0 {
		return nil
	}

	var lastErr error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
