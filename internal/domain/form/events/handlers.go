package form

import (
	"context"
	"fmt"

	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// FormEventHandler handles form-related events
type FormEventHandler struct {
	logger logging.Logger
}

// NewFormEventHandler creates a new form event handler
func NewFormEventHandler(logger logging.Logger) *FormEventHandler {
	return &FormEventHandler{
		logger: logger,
	}
}

// Handle handles form events
func (h *FormEventHandler) Handle(ctx context.Context, event events.Event) error {
	logger := h.logger.With(
		"event_id", event.EventID(),
		"event_type", event.EventType(),
		"timestamp", event.Timestamp(),
	)

	// Log event receipt
	logger.Debug("handling form event")

	// Handle different event types
	switch event.EventType() {
	case string(FormCreatedEventType):
		return h.handleFormCreated(ctx, event)
	case string(FormUpdatedEventType):
		return h.handleFormUpdated(ctx, event)
	case string(FormDeletedEventType):
		return h.handleFormDeleted(ctx, event)
	case string(FormSubmittedEventType):
		return h.handleFormSubmitted(ctx, event)
	case string(FormValidatedEventType):
		return h.handleFormValidated(ctx, event)
	case string(FormProcessedEventType):
		return h.handleFormProcessed(ctx, event)
	case string(FormErrorEventType):
		return h.handleFormError(ctx, event)
	case string(FormStateEventType):
		return h.handleFormState(ctx, event)
	case string(FieldEventType):
		return h.handleFieldEvent(ctx, event)
	case string(AnalyticsEventType):
		return h.handleAnalyticsEvent(ctx, event)
	default:
		logger.Warn("unknown event type")
		return fmt.Errorf("unknown event type: %s", event.EventType())
	}
}

// handleFormCreated handles form creation events
func (h *FormEventHandler) handleFormCreated(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.created")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log form creation with context
	logger.Info("form created",
		"form_id", data["form_id"],
		"user_id", data["user_id"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleFormUpdated handles form update events
func (h *FormEventHandler) handleFormUpdated(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.updated")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log form update with context
	logger.Info("form updated",
		"form_id", data["form_id"],
		"user_id", data["user_id"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleFormDeleted handles form deletion events
func (h *FormEventHandler) handleFormDeleted(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.deleted")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log form deletion with context
	logger.Info("form deleted",
		"form_id", data["form_id"],
		"user_id", data["user_id"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleFormSubmitted handles form submission events
func (h *FormEventHandler) handleFormSubmitted(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.submitted")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log form submission with context
	logger.Info("form submitted",
		"form_id", data["form_id"],
		"submission_id", data["submission_id"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleFormValidated handles form validation events
func (h *FormEventHandler) handleFormValidated(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.validated")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log form validation with context
	logger.Info("form validated",
		"form_id", data["form_id"],
		"is_valid", data["is_valid"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleFormProcessed handles form processing events
func (h *FormEventHandler) handleFormProcessed(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.processed")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log form processing with context
	logger.Info("form processed",
		"form_id", data["form_id"],
		"processing_id", data["processing_id"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleFormError handles form error events
func (h *FormEventHandler) handleFormError(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.error")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log form error with context
	logger.Error("form error",
		"form_id", data["form_id"],
		"error", data["error"],
		"error_type", data["error_type"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleFormState handles form state events
func (h *FormEventHandler) handleFormState(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.state")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log form state change with context
	logger.Info("form state changed",
		"form_id", data["form_id"],
		"state", data["state"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleFieldEvent handles field events
func (h *FormEventHandler) handleFieldEvent(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.field")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log field event with context
	logger.Info("field event",
		"form_id", data["form_id"],
		"field_id", data["field_id"],
		"field_name", data["field_name"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}

// handleAnalyticsEvent handles analytics events
func (h *FormEventHandler) handleAnalyticsEvent(ctx context.Context, event events.Event) error {
	logger := h.logger.With("event_type", "form.analytics")

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Extract event data
	data, ok := event.Data().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	// Log analytics event with context
	logger.Info("analytics event",
		"form_id", data["form_id"],
		"event_type", data["event_type"],
		"user_id", data["user_id"],
		"trace_id", ctx.Value("trace_id"),
	)

	// Add any additional processing here
	return nil
}
