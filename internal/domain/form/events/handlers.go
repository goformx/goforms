package form

import (
	"context"
	"errors"

	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var ErrInvalidEventPayload = errors.New("invalid event payload")

// FormEventHandler handles form-related events
type FormEventHandler struct {
	*events.BaseHandler
}

// NewFormEventHandler creates a new form event handler
func NewFormEventHandler(logger logging.Logger) *FormEventHandler {
	return &FormEventHandler{
		BaseHandler: events.NewBaseHandler(events.HandlerConfig{
			Logger:     logger,
			RetryCount: 3,
		}),
	}
}

// Handle handles form events
func (h *FormEventHandler) Handle(ctx context.Context, event events.Event) error {
	h.LogEvent(event, "info", "handling form event")

	switch event.Name() {
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
		h.LogEvent(event, "warn", "unknown event type")
		return nil
	}
}

// handleFormCreated handles form creation events
func (h *FormEventHandler) handleFormCreated(ctx context.Context, event events.Event) error {
	form, ok := event.Payload().(*model.Form)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "form created", "form_id", form.ID)
	return nil
}

// handleFormUpdated handles form update events
func (h *FormEventHandler) handleFormUpdated(ctx context.Context, event events.Event) error {
	form, ok := event.Payload().(*model.Form)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "form updated", "form_id", form.ID)
	return nil
}

// handleFormDeleted handles form deletion events
func (h *FormEventHandler) handleFormDeleted(ctx context.Context, event events.Event) error {
	formID, ok := event.Payload().(string)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "form deleted", "form_id", formID)
	return nil
}

// handleFormSubmitted handles form submission events
func (h *FormEventHandler) handleFormSubmitted(ctx context.Context, event events.Event) error {
	submission, ok := event.Payload().(*model.FormSubmission)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "form submitted", "form_id", submission.FormID)
	return nil
}

// handleFormValidated handles form validation events
func (h *FormEventHandler) handleFormValidated(ctx context.Context, event events.Event) error {
	payload, ok := event.Payload().(map[string]any)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "form validated", "form_id", payload["form_id"])
	return nil
}

// handleFormProcessed handles form processing events
func (h *FormEventHandler) handleFormProcessed(ctx context.Context, event events.Event) error {
	payload, ok := event.Payload().(map[string]any)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "form processed", "form_id", payload["form_id"])
	return nil
}

// handleFormError handles form error events
func (h *FormEventHandler) handleFormError(ctx context.Context, event events.Event) error {
	payload, ok := event.Payload().(map[string]any)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "error", "form error", "form_id", payload["form_id"], "error", payload["error"])
	return nil
}

// handleFormState handles form state events
func (h *FormEventHandler) handleFormState(ctx context.Context, event events.Event) error {
	payload, ok := event.Payload().(map[string]any)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "form state changed", "form_id", payload["form_id"], "state", payload["state"])
	return nil
}

// handleFieldEvent handles form field events
func (h *FormEventHandler) handleFieldEvent(ctx context.Context, event events.Event) error {
	payload, ok := event.Payload().(map[string]any)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "field event", "form_id", payload["form_id"], "field_id", payload["field_id"])
	return nil
}

// handleAnalyticsEvent handles form analytics events
func (h *FormEventHandler) handleAnalyticsEvent(ctx context.Context, event events.Event) error {
	payload, ok := event.Payload().(map[string]any)
	if !ok {
		return ErrInvalidEventPayload
	}
	h.LogEvent(event, "info", "analytics event", "form_id", payload["form_id"], "event_type", payload["event_type"])
	return nil
}
