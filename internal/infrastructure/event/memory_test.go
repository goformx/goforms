// Package event_test contains tests for the event infrastructure.
package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	commonevents "github.com/goformx/goforms/internal/domain/common/events"
	formevents "github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/event"
	mockevents "github.com/goformx/goforms/test/mocks/events"
	mockform "github.com/goformx/goforms/test/mocks/form"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
)

func TestMemoryPublisher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("NewMemoryPublisher", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger)
		require.NotNil(t, publisher)
		assert.IsType(t, &event.MemoryPublisher{}, publisher)
	})

	t.Run("WithMaxEvents", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		maxEvents := 500
		result := publisher.WithMaxEvents(maxEvents)
		assert.Equal(t, publisher, result)
	})

	t.Run("Publish valid event", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		event := mockform.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()

		err := publisher.Publish(context.Background(), event)
		assert.NoError(t, err)
		assert.Len(t, publisher.GetEvents(), 1)
		assert.Equal(t, event, publisher.GetEvents()[0])
	})

	t.Run("Publish nil event", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		err := publisher.Publish(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, event.ErrInvalidEvent, err)
	})

	t.Run("Publish with handler", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		event := mockform.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()
		handlerCalled := false

		// Subscribe to the event
		err := publisher.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt formevents.Event) error {
			handlerCalled = true
			assert.Equal(t, event, evt)
			return nil
		})
		assert.NoError(t, err)

		// Publish the event
		err = publisher.Publish(context.Background(), event)
		assert.NoError(t, err)

		// Wait for async handler
		time.Sleep(10 * time.Millisecond)
		assert.True(t, handlerCalled)
	})

	t.Run("Publish with handler error", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		event := mockform.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()

		// Subscribe to the event with error
		err := publisher.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt formevents.Event) error {
			return errors.New("handler error")
		})
		assert.NoError(t, err)

		// Publish the event
		err = publisher.Publish(context.Background(), event)
		assert.NoError(t, err)

		// Wait for async handler
		time.Sleep(10 * time.Millisecond)
		// Should not affect the publish operation
		assert.Len(t, publisher.GetEvents(), 1)
	})

	t.Run("Subscribe with nil handler", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		err := publisher.Subscribe(context.Background(), "test.event", nil)
		assert.Error(t, err)
		assert.Equal(t, "handler cannot be nil", err.Error())
	})

	t.Run("Multiple handlers for same event", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		event := mockform.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()
		handler1Called := false
		handler2Called := false

		// Subscribe first handler
		err := publisher.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt formevents.Event) error {
			handler1Called = true
			return nil
		})
		assert.NoError(t, err)

		// Subscribe second handler
		err = publisher.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt formevents.Event) error {
			handler2Called = true
			return nil
		})
		assert.NoError(t, err)

		// Publish the event
		err = publisher.Publish(context.Background(), event)
		assert.NoError(t, err)

		// Wait for async handlers
		time.Sleep(10 * time.Millisecond)
		assert.True(t, handler1Called)
		assert.True(t, handler2Called)
	})

	t.Run("Event overflow", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		publisher.WithMaxEvents(2)

		for i := 0; i < 3; i++ {
			event := mockform.NewMockEvent(ctrl)
			event.EXPECT().Name().Return("test.event").AnyTimes()
			event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
			event.EXPECT().Payload().Return(i).AnyTimes()

			err := publisher.Publish(context.Background(), event)
			assert.NoError(t, err)
		}

		events := publisher.GetEvents()
		assert.Len(t, events, 2)
		assert.Equal(t, 1, events[0].Payload())
		assert.Equal(t, 2, events[1].Payload())
	})

	t.Run("GetEvents returns copy", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		event := mockform.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()

		err := publisher.Publish(context.Background(), event)
		assert.NoError(t, err)

		events1 := publisher.GetEvents()
		events2 := publisher.GetEvents()

		// Should be different slices
		assert.NotSame(t, events1, events2)
		assert.Equal(t, events1, events2)
	})

	t.Run("ClearEvents", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		event := mockform.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()

		err := publisher.Publish(context.Background(), event)
		assert.NoError(t, err)
		assert.Len(t, publisher.GetEvents(), 1)

		publisher.ClearEvents()
		assert.Len(t, publisher.GetEvents(), 0)
	})
}

func TestMemoryEventBus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("NewMemoryEventBus", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger)
		require.NotNil(t, eventBus)
		assert.IsType(t, &event.MemoryEventBus{}, eventBus)
	})

	t.Run("Publish with no subscribers", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)
		event := mockevents.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()
		event.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()

		err := eventBus.Publish(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("Publish with subscriber", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)
		event := mockevents.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()
		event.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()
		handlerCalled := false

		err := eventBus.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt commonevents.Event) error {
			handlerCalled = true
			assert.Equal(t, event, evt)
			return nil
		})
		assert.NoError(t, err)

		err = eventBus.Publish(context.Background(), event)
		assert.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("Publish with handler error", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)
		event := mockevents.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()
		event.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()

		err := eventBus.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt commonevents.Event) error {
			return errors.New("handler error")
		})
		assert.NoError(t, err)

		err = eventBus.Publish(context.Background(), event)
		assert.NoError(t, err)
		// Should not affect the publish operation
	})

	t.Run("Multiple handlers for same event", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)
		event := mockevents.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()
		event.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()
		handler1Called := false
		handler2Called := false

		err := eventBus.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt commonevents.Event) error {
			handler1Called = true
			return nil
		})
		assert.NoError(t, err)

		err = eventBus.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt commonevents.Event) error {
			handler2Called = true
			return nil
		})
		assert.NoError(t, err)

		err = eventBus.Publish(context.Background(), event)
		assert.NoError(t, err)
		assert.True(t, handler1Called)
		assert.True(t, handler2Called)
	})

	t.Run("PublishBatch", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)
		event1 := mockevents.NewMockEvent(ctrl)
		event1.EXPECT().Name().Return("test.event1").AnyTimes()
		event1.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event1.EXPECT().Payload().Return("payload1").AnyTimes()
		event1.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()

		event2 := mockevents.NewMockEvent(ctrl)
		event2.EXPECT().Name().Return("test.event2").AnyTimes()
		event2.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event2.EXPECT().Payload().Return("payload2").AnyTimes()
		event2.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()

		events := []commonevents.Event{event1, event2}
		handlerCalled := 0

		err := eventBus.Subscribe(context.Background(), "test.event1", func(ctx context.Context, evt commonevents.Event) error {
			handlerCalled++
			return nil
		})
		assert.NoError(t, err)

		err = eventBus.Subscribe(context.Background(), "test.event2", func(ctx context.Context, evt commonevents.Event) error {
			handlerCalled++
			return nil
		})
		assert.NoError(t, err)

		err = eventBus.PublishBatch(context.Background(), events)
		assert.NoError(t, err)
		assert.Equal(t, 2, handlerCalled)
	})

	t.Run("PublishBatch with error", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)
		event1 := mockevents.NewMockEvent(ctrl)
		event1.EXPECT().Name().Return("test.event1").AnyTimes()
		event1.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event1.EXPECT().Payload().Return("payload1").AnyTimes()
		event1.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()

		event2 := mockevents.NewMockEvent(ctrl)
		event2.EXPECT().Name().Return("test.event2").AnyTimes()
		event2.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event2.EXPECT().Payload().Return("payload2").AnyTimes()
		event2.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()

		events := []commonevents.Event{event1, event2}

		err := eventBus.Subscribe(context.Background(), "test.event1", func(ctx context.Context, evt commonevents.Event) error {
			return errors.New("handler error")
		})
		assert.NoError(t, err)

		err = eventBus.PublishBatch(context.Background(), events)
		assert.NoError(t, err)
		// Should not affect the batch operation
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)
		event := mockevents.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()
		event.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()
		handlerCalled := false

		err := eventBus.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt commonevents.Event) error {
			handlerCalled = true
			return nil
		})
		assert.NoError(t, err)

		err = eventBus.Unsubscribe(context.Background(), "test.event")
		assert.NoError(t, err)

		err = eventBus.Publish(context.Background(), event)
		assert.NoError(t, err)
		assert.False(t, handlerCalled)
	})

	t.Run("Unsubscribe non-existent event", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)

		err := eventBus.Unsubscribe(context.Background(), "non.existent")
		assert.NoError(t, err)
	})

	t.Run("Start", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)

		err := eventBus.Start(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Stop", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)

		err := eventBus.Stop(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Health", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)

		err := eventBus.Health(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Concurrent access", func(t *testing.T) {
		eventBus := event.NewMemoryEventBus(logger).(*event.MemoryEventBus)
		event := mockevents.NewMockEvent(ctrl)
		event.EXPECT().Name().Return("test.event").AnyTimes()
		event.EXPECT().Timestamp().Return(time.Now()).AnyTimes()
		event.EXPECT().Payload().Return("test payload").AnyTimes()
		event.EXPECT().Metadata().Return(map[string]any{}).AnyTimes()
		handlerCalled := make(chan bool, 10)

		err := eventBus.Subscribe(context.Background(), "test.event", func(ctx context.Context, evt commonevents.Event) error {
			handlerCalled <- true
			return nil
		})
		assert.NoError(t, err)

		// Publish events concurrently
		for i := 0; i < 5; i++ {
			go func() {
				err := eventBus.Publish(context.Background(), event)
				assert.NoError(t, err)
			}()
		}

		// Wait for handlers
		time.Sleep(100 * time.Millisecond)
		close(handlerCalled)

		// Count handler calls
		count := 0
		for range handlerCalled {
			count++
		}
		assert.Equal(t, 5, count)
	})
}

func TestMemoryPublisher_FormEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("Publish form created event", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		form := &model.Form{
			ID:     "form123",
			Title:  "Test Form",
			UserID: "user123",
		}
		event := formevents.NewFormCreatedEvent(form)

		err := publisher.Publish(context.Background(), event)
		assert.NoError(t, err)

		events := publisher.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "form.created", events[0].Name())
		assert.Equal(t, form, events[0].Payload())
	})

	t.Run("Publish form submission event", func(t *testing.T) {
		publisher := event.NewMemoryPublisher(logger).(*event.MemoryPublisher)
		submission := &model.FormSubmission{
			ID:     "submission123",
			FormID: "form123",
			Data:   model.JSON{"name": "John Doe"},
		}
		event := formevents.NewFormSubmissionCreatedEvent(submission)

		err := publisher.Publish(context.Background(), event)
		assert.NoError(t, err)

		events := publisher.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "form.submission.created", events[0].Name())
		assert.Equal(t, submission, events[0].Payload())
	})
}
