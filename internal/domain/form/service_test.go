package form_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/goformx/goforms/internal/domain/common/events"
	domainform "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	mockevents "github.com/goformx/goforms/test/mocks/events"
	mockform "github.com/goformx/goforms/test/mocks/form"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_CreateForm_minimal(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockform.NewMockRepository(ctrl)
	eventBus := mockevents.NewMockEventBus(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	userID := "user123"

	// Create form
	form := model.NewForm(
		userID,
		"Test Form",
		"Test Description",
		model.JSON{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
			},
		},
	)

	// Set up mock expectations in the correct order
	repo.EXPECT().CreateForm(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, f *model.Form) error {
		require.Equal(t, userID, f.UserID)
		require.True(t, f.Active)
		return nil
	})
	eventBus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

	svc := domainform.NewService(repo, eventBus, logger)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	err := svc.CreateForm(ctx, form)
	require.NoError(t, err)
	require.Equal(t, userID, form.UserID)
	require.NotEmpty(t, form.ID)
	require.True(t, form.Active)
}

func TestService_SubmitForm(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockform.NewMockRepository(ctrl)
	eventBus := mockevents.NewMockEventBus(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	// Create test form
	form := model.NewForm(
		"user123",
		"Test Form",
		"Test Description",
		model.JSON{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
				"email": map[string]any{
					"type": "string",
				},
			},
		},
	)

	// Create test submission
	submission := &model.FormSubmission{
		FormID: form.ID,
		Data: model.JSON{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		Status:      model.SubmissionStatusPending,
		SubmittedAt: time.Now(),
	}

	t.Run("successful submission", func(t *testing.T) {
		// Set up mock expectations
		repo.EXPECT().GetFormByID(gomock.Any(), form.ID).Return(form, nil)
		repo.EXPECT().CreateSubmission(
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(func(_ context.Context, s *model.FormSubmission) error {
			require.Equal(t, form.ID, s.FormID)
			require.Equal(t, model.SubmissionStatusPending, s.Status)
			require.NotEmpty(t, s.Data)
			return nil
		})

		// Expect form submitted event
		eventBus.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, event events.Event) error {
			require.Equal(t, "form.submitted", event.Name())
			return nil
		})

		// Expect form validated event
		eventBus.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, event events.Event) error {
			require.Equal(t, "form.validated", event.Name())
			return nil
		})

		// Expect form processed event
		eventBus.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, event events.Event) error {
			require.Equal(t, "form.processed", event.Name())
			return nil
		})

		svc := domainform.NewService(repo, eventBus, logger)
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.SubmitForm(ctx, submission)
		require.NoError(t, err)
	})

	t.Run("form not found", func(t *testing.T) {
		// Set up mock expectations
		repo.EXPECT().GetFormByID(gomock.Any(), form.ID).Return(nil, nil)

		svc := domainform.NewService(repo, eventBus, logger)
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.SubmitForm(ctx, submission)
		require.Error(t, err)
		require.Equal(t, "form not found", err.Error())
	})

	t.Run("invalid submission data", func(t *testing.T) {
		invalidSubmission := &model.FormSubmission{
			FormID: form.ID,
			Data:   nil, // Missing required data
		}

		svc := domainform.NewService(repo, eventBus, logger)
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.SubmitForm(ctx, invalidSubmission)
		require.Error(t, err)
		require.Contains(t, err.Error(), "submission data is required")
	})

	t.Run("repository error", func(t *testing.T) {
		// Set up mock expectations
		repo.EXPECT().GetFormByID(gomock.Any(), form.ID).Return(form, nil)
		repo.EXPECT().CreateSubmission(gomock.Any(), gomock.Any()).Return(errors.New("database error"))

		svc := domainform.NewService(repo, eventBus, logger)
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.SubmitForm(ctx, submission)
		require.Error(t, err)
		require.Equal(t, "database error", err.Error())
	})
}
