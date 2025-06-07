package user

import (
	"context"
	"errors"
	"strconv"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"gorm.io/gorm"
)

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
)

// Store implements user.Repository interface
type Store struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewStore creates a new user store
func NewStore(db *database.GormDB, logger logging.Logger) user.Repository {
	logger.Debug("creating user store",
		logging.BoolField("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create stores a new user
func (s *Store) Create(ctx context.Context, u *user.User) error {
	result := s.db.WithContext(ctx).Create(u)
	if result.Error != nil {
		return domainerrors.Wrap(result.Error, domainerrors.ErrCodeServerError, "failed to insert user")
	}
	return nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	result := s.db.WithContext(ctx).Where("email = ?", email).First(&u)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
		}
		return nil, domainerrors.Wrap(result.Error, domainerrors.ErrCodeServerError, "failed to get user")
	}
	return &u, nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	result := s.db.WithContext(ctx).First(&u, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
		}
		return nil, domainerrors.Wrap(result.Error, domainerrors.ErrCodeServerError, "failed to get user")
	}
	return &u, nil
}

// GetByIDString retrieves a user by ID string
func (s *Store) GetByIDString(ctx context.Context, id string) (*user.User, error) {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, user.ErrInvalidUserID
	}
	return s.GetByID(ctx, uint(userID))
}

// Update updates a user
func (s *Store) Update(ctx context.Context, userModel *user.User) error {
	result := s.db.WithContext(ctx).Save(userModel)
	if result.Error != nil {
		return domainerrors.Wrap(result.Error, domainerrors.ErrCodeServerError, "failed to update user")
	}
	if result.RowsAffected == 0 {
		return domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
	}
	return nil
}

// Delete removes a user by ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&user.User{}, id)
	if result.Error != nil {
		return domainerrors.Wrap(result.Error, domainerrors.ErrCodeServerError, "failed to delete user")
	}
	if result.RowsAffected == 0 {
		return domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
	}
	return nil
}

// List returns all users
func (s *Store) List(ctx context.Context) ([]user.User, error) {
	var users []user.User
	result := s.db.WithContext(ctx).Order("id").Find(&users)
	if result.Error != nil {
		return nil, domainerrors.Wrap(result.Error, domainerrors.ErrCodeServerError, "failed to list users")
	}
	return users, nil
}

// ListPaginated returns a paginated list of users
func (s *Store) ListPaginated(ctx context.Context, offset, limit int) ([]*user.User, error) {
	var users []*user.User
	result := s.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id").Find(&users)
	if result.Error != nil {
		return nil, domainerrors.Wrap(result.Error, domainerrors.ErrCodeServerError, "failed to list users")
	}
	return users, nil
}

// Count returns the total number of users
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int64
	result := s.db.WithContext(ctx).Model(&user.User{}).Count(&count)
	if result.Error != nil {
		return 0, domainerrors.Wrap(result.Error, domainerrors.ErrCodeServerError, "failed to count users")
	}
	return int(count), nil
}
