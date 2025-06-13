package form_test

import (
	"context"
	"testing"
	"time"

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
