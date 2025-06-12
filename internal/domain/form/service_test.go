package form

import (
	"context"
	"testing"
	"time"

	"github.com/goformx/goforms/internal/domain/form/model"
	mockform "github.com/goformx/goforms/test/mocks/form"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_CreateForm_minimal(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockform.NewMockRepository(ctrl)
	publisher := mockform.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	userID := "user123"
	form := model.NewForm(
		userID,
		"Test Form",
		"A test form description",
		model.JSON{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
			},
		},
	)

	logger.EXPECT().WithUserID(userID).Return(logger)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, f *model.Form) error {
		require.NotEmpty(t, f.ID)
		require.Equal(t, userID, f.UserID)
		require.True(t, f.Active)
		return nil
	})
	publisher.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

	svc := NewService(repo, publisher, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := svc.CreateForm(ctx, userID, form)
	require.NoError(t, err)
	require.Equal(t, userID, form.UserID)
	require.NotEmpty(t, form.ID)
	require.True(t, form.Active)
}
