package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// DefaultTimeout is the default timeout for event handlers
	DefaultTimeout    = 30 * time.Second
	DefaultRetryCount = 3
	DefaultRetryDelay = time.Second
	DefaultMaxBackoff = 30 * time.Second
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
		config.Timeout = 30 * time.Second
	}
	return &BaseHandler{
		config: config,
	}
}

// HandleWithRetry handles an event with retry logic
func (h *BaseHandler) HandleWithRetry(ctx context.Context, event Event, handler func(ctx context.Context, event Event) error) error {
	var lastErr error
	for i := range h.config.RetryCount {
		if err := handler(ctx, event); err != nil {
			lastErr = err
			log.Printf("Retry %d/%d failed: %v", i+1, h.config.RetryCount, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		return nil
	}
	return fmt.Errorf("event handling failed after %d attempts: %w", h.config.RetryCount, lastErr)
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
