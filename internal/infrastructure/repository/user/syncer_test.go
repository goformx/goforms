package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
	mockuser "github.com/goformx/goforms/test/mocks/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSyncer_EnsureUser_userExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := NewLaravelUserSyncer(repo)

	ctx := context.Background()
	userID := "42"
	existing := &entities.User{ID: userID, Email: "existing@example.com"}

	repo.EXPECT().
		GetByID(ctx, userID).
		Return(existing, nil)

	err := syncer.EnsureUser(ctx, userID)
	require.NoError(t, err)
}

func TestSyncer_EnsureUser_userNotFound_createsShadow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := NewLaravelUserSyncer(repo)

	ctx := context.Background()
	userID := "1"
	notFoundErr := fmt.Errorf("get user by ID: %w", common.NewNotFoundError("get_by_id", "user", userID))

	repo.EXPECT().
		GetByID(ctx, userID).
		Return(nil, notFoundErr)
	repo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, u *entities.User) error {
			assert.Equal(t, userID, u.ID)
			assert.Equal(t, "laravel-1@localhost", u.Email)
			assert.Equal(t, "Laravel", u.FirstName)
			assert.Equal(t, "Sync", u.LastName)
			assert.Equal(t, entities.LaravelShadowPassword, u.HashedPassword)
			return nil
		})

	err := syncer.EnsureUser(ctx, userID)
	require.NoError(t, err)
}

func TestSyncer_EnsureUser_getByIDOtherError_returnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := NewLaravelUserSyncer(repo)

	ctx := context.Background()
	userID := "1"
	dbErr := errors.New("database connection failed")

	repo.EXPECT().
		GetByID(ctx, userID).
		Return(nil, dbErr)

	err := syncer.EnsureUser(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get user by ID")
}

func TestSyncer_EnsureUser_createFails_returnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := NewLaravelUserSyncer(repo)

	ctx := context.Background()
	userID := "1"
	notFoundErr := fmt.Errorf("get user by ID: %w", common.NewNotFoundError("get_by_id", "user", userID))
	createErr := errors.New("unique constraint violation")

	repo.EXPECT().
		GetByID(ctx, userID).
		Return(nil, notFoundErr)
	repo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(createErr)

	err := syncer.EnsureUser(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create Laravel shadow user")
}
