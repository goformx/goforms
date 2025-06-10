package user

import (
	"context"
	"errors"
	"strconv"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/persistence/store/common"
	"gorm.io/gorm"
)

// Store implements user.Repository interface
type Store struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewStore creates a new user store
func NewStore(db *database.GormDB, logger logging.Logger) user.Repository {
	logger.Debug("user store initialized", "service", "user")
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create stores a new user
func (s *Store) Create(ctx context.Context, u *user.User) error {
	result := s.db.WithContext(ctx).Create(u)
	if result.Error != nil {
		return common.NewDatabaseError("create", "user", strconv.FormatUint(uint64(u.ID), 10), result.Error)
	}
	return nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	result := s.db.WithContext(ctx).Where("email = ?", email).First(&u)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFoundError("get_by_email", "user", email)
		}
		return nil, common.NewDatabaseError("get_by_email", "user", email, result.Error)
	}
	return &u, nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	result := s.db.WithContext(ctx).First(&u, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFoundError("get_by_id", "user", strconv.FormatUint(uint64(id), 10))
		}
		return nil, common.NewDatabaseError("get_by_id", "user", strconv.FormatUint(uint64(id), 10), result.Error)
	}
	return &u, nil
}

// GetByIDString retrieves a user by ID string
func (s *Store) GetByIDString(ctx context.Context, id string) (*user.User, error) {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, common.NewInvalidInputError("get_by_id_string", "user", id, err)
	}
	return s.GetByID(ctx, uint(userID))
}

// Update updates a user
func (s *Store) Update(ctx context.Context, userModel *user.User) error {
	result := s.db.WithContext(ctx).Save(userModel)
	if result.Error != nil {
		return common.NewDatabaseError("update", "user", strconv.FormatUint(uint64(userModel.ID), 10), result.Error)
	}
	if result.RowsAffected == 0 {
		return common.NewNotFoundError("update", "user", strconv.FormatUint(uint64(userModel.ID), 10))
	}
	return nil
}

// Delete removes a user by ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&user.User{}, id)
	if result.Error != nil {
		return common.NewDatabaseError("delete", "user", strconv.FormatUint(uint64(id), 10), result.Error)
	}
	if result.RowsAffected == 0 {
		return common.NewNotFoundError("delete", "user", strconv.FormatUint(uint64(id), 10))
	}
	return nil
}

// List returns all users
func (s *Store) List(ctx context.Context) ([]user.User, error) {
	var users []user.User
	result := s.db.WithContext(ctx).Order("id").Find(&users)
	if result.Error != nil {
		return nil, common.NewDatabaseError("list", "user", "", result.Error)
	}
	return users, nil
}

// ListPaginated returns a paginated list of users
func (s *Store) ListPaginated(ctx context.Context, params common.PaginationParams) common.PaginationResult {
	var users []user.User
	var total int64

	// Get total count
	if err := s.db.WithContext(ctx).Model(&user.User{}).Count(&total).Error; err != nil {
		return common.PaginationResult{
			Items:      nil,
			TotalItems: 0,
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: 0,
		}
	}

	// Get paginated results
	result := s.db.WithContext(ctx).
		Order("id").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&users)

	if result.Error != nil {
		return common.PaginationResult{
			Items:      nil,
			TotalItems: 0,
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: 0,
		}
	}

	return common.NewPaginationResult(users, int(total), params.Page, params.PageSize)
}

// Count returns the total number of users
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int64
	result := s.db.WithContext(ctx).Model(&user.User{}).Count(&count)
	if result.Error != nil {
		return 0, common.NewDatabaseError("count", "user", "", result.Error)
	}
	return int(count), nil
}
