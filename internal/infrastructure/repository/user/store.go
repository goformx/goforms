// Package repository provides the user repository implementation
package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
)

// Store implements user.Repository interface
type Store struct {
	db     database.DB
	logger logging.Logger
}

// NewStore creates a new user store
func NewStore(db database.DB, logger logging.Logger) user.Repository {
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create stores a new user
func (s *Store) Create(ctx context.Context, u *entities.User) error {
	result := s.db.GetDB().WithContext(ctx).Create(u)
	if result.Error != nil {
		dbErr := common.NewDatabaseError("create", "user", u.ID, result.Error)

		return fmt.Errorf("create user: %w", dbErr)
	}

	return nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var u entities.User

	result := s.db.GetDB().WithContext(ctx).Where("email = ?", email).First(&u)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			notFoundErr := common.NewNotFoundError("get_by_email", "user", email)

			return nil, fmt.Errorf("get user by email: %w", notFoundErr)
		}

		dbErr := common.NewDatabaseError("get_by_email", "user", email, result.Error)

		return nil, fmt.Errorf("get user by email: %w", dbErr)
	}

	return &u, nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var u entities.User

	result := s.db.GetDB().WithContext(ctx).First(&u, "uuid = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			notFoundErr := common.NewNotFoundError("get_by_id", "user", id)

			return nil, fmt.Errorf("get user by ID: %w", notFoundErr)
		}

		dbErr := common.NewDatabaseError("get_by_id", "user", id, result.Error)

		return nil, fmt.Errorf("get user by ID: %w", dbErr)
	}

	return &u, nil
}

// GetByIDString retrieves a user by ID string
func (s *Store) GetByIDString(ctx context.Context, id string) (*entities.User, error) {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		invalidErr := common.NewInvalidInputError("get_by_id_string", "user", id, err)

		return nil, fmt.Errorf("get user by ID string: %w", invalidErr)
	}

	return s.GetByID(ctx, strconv.FormatUint(userID, 10))
}

// Update updates a user
func (s *Store) Update(ctx context.Context, userModel *entities.User) error {
	result := s.db.GetDB().WithContext(ctx).Save(userModel)
	if result.Error != nil {
		dbErr := common.NewDatabaseError("update", "user", userModel.ID, result.Error)

		return fmt.Errorf("update user: %w", dbErr)
	}

	if result.RowsAffected == 0 {
		notFoundErr := common.NewNotFoundError("update", "user", userModel.ID)

		return fmt.Errorf("update user: %w", notFoundErr)
	}

	return nil
}

// Delete removes a user by ID
func (s *Store) Delete(ctx context.Context, id string) error {
	result := s.db.GetDB().WithContext(ctx).Delete(&entities.User{}, "uuid = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", common.NewDatabaseError("delete", "user", id, result.Error))
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete user: %w", common.NewNotFoundError("delete", "user", id))
	}

	return nil
}

// List returns all users
func (s *Store) List(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var users []*entities.User

	result := s.db.GetDB().WithContext(ctx).Order("uuid").Offset(offset).Limit(limit).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("list users: %w", common.NewDatabaseError("list", "user", "", result.Error))
	}

	return users, nil
}

// ListPaginated returns a paginated list of users
func (s *Store) ListPaginated(ctx context.Context, params common.PaginationParams) common.PaginationResult {
	var users []*entities.User

	var total int64

	// Get total count
	if err := s.db.GetDB().WithContext(ctx).Model(&entities.User{}).Count(&total).Error; err != nil {
		return common.PaginationResult{
			Items:      nil,
			TotalItems: 0,
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: 0,
		}
	}

	// Get paginated results
	result := s.db.GetDB().WithContext(ctx).
		Order("uuid").
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

	result := s.db.GetDB().WithContext(ctx).Model(&entities.User{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("count users: %w", common.NewDatabaseError("count", "user", "", result.Error))
	}

	return int(count), nil
}

// GetByUsername retrieves a user by username
func (s *Store) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	var u entities.User

	result := s.db.GetDB().WithContext(ctx).Where("username = ?", username).First(&u)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get user by username: %w",
				common.NewNotFoundError("get_by_username", "user", username))
		}

		return nil, fmt.Errorf("get user by username: %w",
			common.NewDatabaseError("get_by_username", "user", username, result.Error))
	}

	return &u, nil
}

// GetByRole retrieves users by role
func (s *Store) GetByRole(ctx context.Context, role string, offset, limit int) ([]*entities.User, error) {
	var users []*entities.User

	result := s.db.GetDB().WithContext(ctx).
		Where("role = ?", role).
		Order("uuid").
		Offset(offset).
		Limit(limit).
		Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("get users by role: %w", common.NewDatabaseError("get_by_role", "user", role, result.Error))
	}

	return users, nil
}

// GetActiveUsers retrieves all active users
func (s *Store) GetActiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var users []*entities.User

	result := s.db.GetDB().WithContext(ctx).
		Where("active = ?", true).
		Order("uuid").
		Offset(offset).
		Limit(limit).
		Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("get active users: %w", common.NewDatabaseError("get_active_users", "user", "", result.Error))
	}

	return users, nil
}

// GetInactiveUsers retrieves all inactive users
func (s *Store) GetInactiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var users []*entities.User

	result := s.db.GetDB().WithContext(ctx).
		Where("active = ?", false).
		Order("uuid").
		Offset(offset).
		Limit(limit).
		Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("get inactive users: %w",
			common.NewDatabaseError("get_inactive_users", "user", "", result.Error))
	}

	return users, nil
}

// Search searches users by name or email
func (s *Store) Search(ctx context.Context, query string, offset, limit int) ([]*entities.User, error) {
	var users []*entities.User

	result := s.db.GetDB().WithContext(ctx).
		Where("name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%").
		Order("uuid").
		Offset(offset).
		Limit(limit).
		Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("search users: %w", common.NewDatabaseError("search", "user", query, result.Error))
	}

	return users, nil
}
