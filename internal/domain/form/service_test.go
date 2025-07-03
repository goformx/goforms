package form_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/domain/common/events"
	domainform "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	mockevents "github.com/goformx/goforms/test/mocks/events"
	mockform "github.com/goformx/goforms/test/mocks/form"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
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

func TestService_ListForms(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockform.NewMockRepository(ctrl)
	eventBus := mockevents.NewMockEventBus(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	userID := "user123"
	expectedForms := []*model.Form{
		{
			ID:          "form1",
			UserID:      userID,
			Title:       "Test Form 1",
			Description: "Description 1",
			Status:      "draft",
			Active:      true,
		},
		{
			ID:          "form2",
			UserID:      userID,
			Title:       "Test Form 2",
			Description: "Description 2",
			Status:      "published",
			Active:      true,
		},
	}

	t.Run("successful list", func(t *testing.T) {
		repo.EXPECT().ListForms(gomock.Any(), userID).Return(expectedForms, nil)

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		forms, err := svc.ListForms(ctx, userID)
		require.NoError(t, err)
		require.Len(t, forms, 2)
		require.Equal(t, expectedForms[0].ID, forms[0].ID)
		require.Equal(t, expectedForms[1].ID, forms[1].ID)
	})

	t.Run("repository error", func(t *testing.T) {
		repo.EXPECT().ListForms(gomock.Any(), userID).Return(nil, errors.New("database error"))

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		forms, err := svc.ListForms(ctx, userID)
		require.Error(t, err)
		require.Nil(t, forms)
		require.Contains(t, err.Error(), "failed to list forms")
	})

	t.Run("empty list", func(t *testing.T) {
		repo.EXPECT().ListForms(gomock.Any(), userID).Return([]*model.Form{}, nil)

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		forms, err := svc.ListForms(ctx, userID)
		require.NoError(t, err)
		require.Empty(t, forms)
	})
}

func TestService_UpdateForm(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockform.NewMockRepository(ctrl)
	eventBus := mockevents.NewMockEventBus(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	form := &model.Form{
		ID:          "form123",
		UserID:      "user123",
		Title:       "Original Title",
		Description: "Original Description",
		Status:      "draft",
		Active:      true,
		Schema: model.JSON{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
			},
		},
	}

	t.Run("successful update", func(t *testing.T) {
		// Update form fields
		form.Title = "Updated Title"
		form.Description = "Updated Description"
		form.Status = "published"

		repo.EXPECT().UpdateForm(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, f *model.Form) error {
			require.Equal(t, "Updated Title", f.Title)
			require.Equal(t, "Updated Description", f.Description)
			require.Equal(t, "published", f.Status)

			return nil
		})

		eventBus.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, event events.Event) error {
			require.Equal(t, "form.updated", event.Name())

			return nil
		})

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.UpdateForm(ctx, form)
		require.NoError(t, err)
	})

	t.Run("validation error", func(t *testing.T) {
		invalidForm := &model.Form{
			ID:     "form123",
			UserID: "user123",
			Title:  "", // Invalid: empty title
		}

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.UpdateForm(ctx, invalidForm)
		require.Error(t, err)
		require.Contains(t, err.Error(), "validate form")
	})

	t.Run("repository error", func(t *testing.T) {
		repo.EXPECT().UpdateForm(gomock.Any(), gomock.Any()).Return(errors.New("database error"))

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.UpdateForm(ctx, form)
		require.Error(t, err)
		require.Contains(t, err.Error(), "update form in repository")
	})

	t.Run("event bus error", func(t *testing.T) {
		repo.EXPECT().UpdateForm(gomock.Any(), gomock.Any()).Return(nil)
		eventBus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(errors.New("event bus error"))
		logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Return()

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.UpdateForm(ctx, form)
		require.NoError(t, err) // Event bus errors are logged but don't fail the operation
	})
}

func TestService_DeleteForm(t *testing.T) {
	formID := "form123"

	t.Run("successful deletion", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mockform.NewMockRepository(ctrl)
		eventBus := mockevents.NewMockEventBus(ctrl)
		logger := mocklogging.NewMockLogger(ctrl)

		repo.EXPECT().DeleteForm(gomock.Any(), formID).Return(nil)
		eventBus.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, event events.Event) error {
			require.Equal(t, "form.deleted", event.Name())

			return nil
		})

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.DeleteForm(ctx, formID)
		require.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mockform.NewMockRepository(ctrl)
		eventBus := mockevents.NewMockEventBus(ctrl)
		logger := mocklogging.NewMockLogger(ctrl)

		repo.EXPECT().DeleteForm(gomock.Any(), formID).Return(errors.New("database error"))

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.DeleteForm(ctx, formID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete form")
	})

	t.Run("event bus error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mockform.NewMockRepository(ctrl)
		eventBus := mockevents.NewMockEventBus(ctrl)
		logger := mocklogging.NewMockLogger(ctrl)

		repo.EXPECT().DeleteForm(gomock.Any(), formID).Return(nil)
		eventBus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(errors.New("event bus error"))
		logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Return()

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.DeleteForm(ctx, formID)
		require.NoError(t, err) // Event bus errors are logged but don't fail the operation
	})

	t.Run("empty form ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mockform.NewMockRepository(ctrl)
		eventBus := mockevents.NewMockEventBus(ctrl)
		logger := mocklogging.NewMockLogger(ctrl)

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := svc.DeleteForm(ctx, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete form")
	})
}

func TestService_GetForm(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockform.NewMockRepository(ctrl)
	eventBus := mockevents.NewMockEventBus(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	expectedForm := &model.Form{
		ID:          "form123",
		UserID:      "user123",
		Title:       "Test Form",
		Description: "Test Description",
		Status:      "draft",
		Active:      true,
	}

	t.Run("successful get", func(t *testing.T) {
		repo.EXPECT().GetFormByID(gomock.Any(), "form123").Return(expectedForm, nil)

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		form, err := svc.GetForm(ctx, "form123")
		require.NoError(t, err)
		require.Equal(t, expectedForm, form)
	})

	t.Run("form not found", func(t *testing.T) {
		repo.EXPECT().GetFormByID(gomock.Any(), "nonexistent").Return(nil, errors.New("not found"))

		svc := domainform.NewService(repo, eventBus, logger)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		form, err := svc.GetForm(ctx, "nonexistent")
		require.Error(t, err)
		require.Nil(t, form)
		require.Contains(t, err.Error(), "get form by ID")
	})
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
		require.Equal(t, "create form submission: database error", err.Error())
	})
}
