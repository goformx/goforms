// Package events provides event handling infrastructure for the domain layer.
// It includes interfaces and implementations for event publishing and handling.
package events

import (
	"context"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// DefaultTimeout is the default timeout for event handlers
	DefaultTimeout = 30 * time.Second
	// DefaultRetryCount is the default number of retry attempts for event handlers
	DefaultRetryCount = 3
	// DefaultRetryDelay is the default delay between retry attempts for event handlers
	DefaultRetryDelay = time.Second
	// DefaultMaxBackoff is the maximum backoff duration for retries
	DefaultMaxBackoff = 30 * time.Second
	// DefaultRetryTimeout is the default timeout for retry operations
	DefaultRetryTimeout = 30 * time.Second
)

// HandlerConfig represents the configuration for an event handler
type HandlerConfig struct {
	Logger     logging.Logger
	RetryCount int
	Timeout    time.Duration
	RetryDelay time.Duration
	MaxBackoff time.Duration
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
		config.Timeout = DefaultTimeout
	}

	return &BaseHandler{
		config: config,
	}
}

// HandleWithRetry handles an event with retry logic
func (h *BaseHandler) HandleWithRetry(
	ctx context.Context,
	event Event,
	handler func(ctx context.Context, event Event) error,
) error {
	config := h.config
	config.Timeout = DefaultRetryTimeout

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	var lastErr error

	for i := range h.config.RetryCount {
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry timeout after %d attempts: %w", i, lastErr)
		default:
			if err := handler(ctx, event); err != nil {
				lastErr = err

				time.Sleep(h.config.RetryDelay)

				continue
			}

			return nil
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// LogEvent logs an event
func (h *BaseHandler) LogEvent(event Event, level, message string, fields ...any) {
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

// NewHandlerConfig creates a new handler configuration with default values
func NewHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		Timeout:    DefaultTimeout,
		RetryCount: DefaultRetryCount,
		RetryDelay: DefaultRetryDelay,
		MaxBackoff: DefaultMaxBackoff,
	}
}
