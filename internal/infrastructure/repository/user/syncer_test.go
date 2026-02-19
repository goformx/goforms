package repository_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
	repository "github.com/goformx/goforms/internal/infrastructure/repository/user"
	mockuser "github.com/goformx/goforms/test/mocks/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestSyncer_EnsureUser_emptyUserID_returnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := repository.NewUserEnsurer(repo)

	err := syncer.EnsureUser(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user ID must not be empty")
}

func TestSyncer_EnsureUser_userExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := repository.NewUserEnsurer(repo)

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
	syncer := repository.NewUserEnsurer(repo)

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
			assert.Equal(t, "laravel-1@shadow.local", u.Email)
			assert.Equal(t, "Laravel", u.FirstName)
			assert.Equal(t, "Sync", u.LastName)
			assert.Equal(t, "!shadow-no-login", u.HashedPassword)
			return nil
		})

	err := syncer.EnsureUser(ctx, userID)
	require.NoError(t, err)
}

func TestSyncer_EnsureUser_getByIDOtherError_returnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := repository.NewUserEnsurer(repo)

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

func TestSyncer_EnsureUser_createFails_retryGetByIDFails_returnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := repository.NewUserEnsurer(repo)

	ctx := context.Background()
	userID := "1"
	notFoundErr := fmt.Errorf("get user by ID: %w", common.NewNotFoundError("get_by_id", "user", userID))
	createErr := errors.New("database write error")

	// First GetByID: not found. Second GetByID (retry after Create fails): still not found.
	// gomock consumes same-method expectations in FIFO order.
	repo.EXPECT().GetByID(ctx, userID).Return(nil, notFoundErr)
	repo.EXPECT().GetByID(ctx, userID).Return(nil, notFoundErr)
	repo.EXPECT().Create(ctx, gomock.Any()).Return(createErr)

	err := syncer.EnsureUser(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create shadow user")
}

func TestSyncer_EnsureUser_createRace_retryGetByIDSucceeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := repository.NewUserEnsurer(repo)

	ctx := context.Background()
	userID := "1"
	notFoundErr := fmt.Errorf("get user by ID: %w", common.NewNotFoundError("get_by_id", "user", userID))
	existing := &entities.User{ID: userID, Email: "laravel-1@shadow.local"}

	// First GetByID: not found. Second GetByID (retry after Create race): user exists.
	repo.EXPECT().GetByID(ctx, userID).Return(nil, notFoundErr)
	repo.EXPECT().GetByID(ctx, userID).Return(existing, nil)
	repo.EXPECT().Create(ctx, gomock.Any()).Return(errors.New("unique constraint violation"))

	err := syncer.EnsureUser(ctx, userID)
	require.NoError(t, err)
}

func TestSyncer_shadowUser_emailTruncatedForLongID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockuser.NewMockRepository(ctrl)
	syncer := repository.NewUserEnsurer(repo)

	ctx := context.Background()
	longID := strings.Repeat("x", 300)
	notFoundErr := fmt.Errorf("get user by ID: %w", common.NewNotFoundError("get_by_id", "user", longID))

	repo.EXPECT().GetByID(ctx, longID).Return(nil, notFoundErr)
	repo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, u *entities.User) error {
			assert.LessOrEqual(t, len(u.Email), 255)
			assert.True(t, strings.HasPrefix(u.Email, "laravel-"))
			return nil
		})

	err := syncer.EnsureUser(ctx, longID)
	require.NoError(t, err)
}

func TestSyncer_shadowUser_passwordCanNeverMatchBcrypt(t *testing.T) {
	// The shadow password "!shadow-no-login" is not a valid bcrypt hash,
	// so bcrypt.CompareHashAndPassword will always reject it.
	shadowPwd := "!shadow-no-login"
	require.Error(t, bcrypt.CompareHashAndPassword([]byte(shadowPwd), []byte("anything")))
	require.Error(t, bcrypt.CompareHashAndPassword([]byte(shadowPwd), []byte("")))
	require.Error(t, bcrypt.CompareHashAndPassword([]byte(shadowPwd), []byte(shadowPwd)))
}
