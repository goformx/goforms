// Package repository provides the user repository implementation
package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
	"github.com/google/uuid"
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

// UserModel is the infrastructure representation of a user for GORM
// This struct contains all GORM-specific fields and tags
// and is mapped to/from the pure domain User entity

type UserModel struct {
	ID             string         `gorm:"column:uuid;primaryKey;type:uuid;default:gen_random_uuid()"`
	Email          string         `gorm:"uniqueIndex;not null;size:255"`
	HashedPassword string         `gorm:"column:hashed_password;not null;size:255"`
	FirstName      string         `gorm:"not null;size:100"`
	LastName       string         `gorm:"not null;size:100"`
	Role           string         `gorm:"not null;size:50;default:user"`
	Active         bool           `gorm:"not null;default:true"`
	CreatedAt      time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string { return "users" }

func (m *UserModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.Role == "" {
		m.Role = "user"
	}
	if !m.Active {
		m.Active = true
	}
	return nil
}

func (m *UserModel) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

func (m *UserModel) AfterFind(tx *gorm.DB) error {
	if m.ID != "" {
		if _, err := uuid.Parse(m.ID); err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
	}
	return nil
}

// Mapper: domain <-> infra
func userModelFromDomain(u *entities.User) *UserModel {
	if u == nil {
		return nil
	}
	return &UserModel{
		ID:             u.ID,
		Email:          u.Email,
		HashedPassword: u.HashedPassword,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Role:           u.Role,
		Active:         u.Active,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

func (m *UserModel) ToDomain() *entities.User {
	if m == nil {
		return nil
	}
	return &entities.User{
		ID:             m.ID,
		Email:          m.Email,
		HashedPassword: m.HashedPassword,
		FirstName:      m.FirstName,
		LastName:       m.LastName,
		Role:           m.Role,
		Active:         m.Active,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// Create stores a new user
func (s *Store) Create(ctx context.Context, u *entities.User) error {
	userModel := userModelFromDomain(u)
	result := s.db.GetDB().WithContext(ctx).Create(userModel)
	if result.Error != nil {
		dbErr := common.NewDatabaseError("create", "user", u.ID, result.Error)

		return fmt.Errorf("create user: %w", dbErr)
	}

	return nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var userModel UserModel

	result := s.db.GetDB().WithContext(ctx).Where("email = ?", email).First(&userModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			notFoundErr := common.NewNotFoundError("get_by_email", "user", email)

			return nil, fmt.Errorf("get user by email: %w", notFoundErr)
		}

		dbErr := common.NewDatabaseError("get_by_email", "user", email, result.Error)

		return nil, fmt.Errorf("get user by email: %w", dbErr)
	}

	return userModel.ToDomain(), nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var userModel UserModel

	result := s.db.GetDB().WithContext(ctx).First(&userModel, "uuid = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			notFoundErr := common.NewNotFoundError("get_by_id", "user", id)

			return nil, fmt.Errorf("get user by ID: %w", notFoundErr)
		}

		dbErr := common.NewDatabaseError("get_by_id", "user", id, result.Error)

		return nil, fmt.Errorf("get user by ID: %w", dbErr)
	}

	return userModel.ToDomain(), nil
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
func (s *Store) Update(ctx context.Context, userEntity *entities.User) error {
	userModel := userModelFromDomain(userEntity)
	result := s.db.GetDB().WithContext(ctx).Save(userModel)
	if result.Error != nil {
		dbErr := common.NewDatabaseError("update", "user", userEntity.ID, result.Error)

		return fmt.Errorf("update user: %w", dbErr)
	}

	if result.RowsAffected == 0 {
		notFoundErr := common.NewNotFoundError("update", "user", userEntity.ID)

		return fmt.Errorf("update user: %w", notFoundErr)
	}

	return nil
}

// Delete removes a user by ID
func (s *Store) Delete(ctx context.Context, id string) error {
	result := s.db.GetDB().WithContext(ctx).Delete(&UserModel{}, "uuid = ?", id)
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
	var userModels []*UserModel

	result := s.db.GetDB().WithContext(ctx).Order("uuid").Offset(offset).Limit(limit).Find(&userModels)
	if result.Error != nil {
		return nil, fmt.Errorf("list users: %w", common.NewDatabaseError("list", "user", "", result.Error))
	}

	users := make([]*entities.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = userModel.ToDomain()
	}

	return users, nil
}

// ListPaginated returns a paginated list of users
func (s *Store) ListPaginated(ctx context.Context, params common.PaginationParams) common.PaginationResult {
	var userModels []*UserModel

	var total int64

	// Get total count
	if err := s.db.GetDB().WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
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
		Find(&userModels)

	if result.Error != nil {
		return common.PaginationResult{
			Items:      nil,
			TotalItems: 0,
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: 0,
		}
	}

	users := make([]*entities.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = userModel.ToDomain()
	}

	return common.NewPaginationResult(users, int(total), params.Page, params.PageSize)
}

// Count returns the total number of users
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int64

	result := s.db.GetDB().WithContext(ctx).Model(&UserModel{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("count users: %w", common.NewDatabaseError("count", "user", "", result.Error))
	}

	return int(count), nil
}

// GetByUsername retrieves a user by username
func (s *Store) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	var userModel UserModel

	result := s.db.GetDB().WithContext(ctx).Where("username = ?", username).First(&userModel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get user by username: %w",
				common.NewNotFoundError("get_by_username", "user", username))
		}

		return nil, fmt.Errorf("get user by username: %w",
			common.NewDatabaseError("get_by_username", "user", username, result.Error))
	}

	return userModel.ToDomain(), nil
}

// GetByRole retrieves users by role
func (s *Store) GetByRole(ctx context.Context, role string, offset, limit int) ([]*entities.User, error) {
	var userModels []*UserModel

	result := s.db.GetDB().WithContext(ctx).
		Where("role = ?", role).
		Order("uuid").
		Offset(offset).
		Limit(limit).
		Find(&userModels)
	if result.Error != nil {
		return nil, fmt.Errorf("get users by role: %w", common.NewDatabaseError("get_by_role", "user", role, result.Error))
	}

	users := make([]*entities.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = userModel.ToDomain()
	}

	return users, nil
}

// GetActiveUsers retrieves all active users
func (s *Store) GetActiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var userModels []*UserModel

	result := s.db.GetDB().WithContext(ctx).
		Where("active = ?", true).
		Order("uuid").
		Offset(offset).
		Limit(limit).
		Find(&userModels)
	if result.Error != nil {
		return nil, fmt.Errorf("get active users: %w", common.NewDatabaseError("get_active_users", "user", "", result.Error))
	}

	users := make([]*entities.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = userModel.ToDomain()
	}

	return users, nil
}

// GetInactiveUsers retrieves all inactive users
func (s *Store) GetInactiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var userModels []*UserModel

	result := s.db.GetDB().WithContext(ctx).
		Where("active = ?", false).
		Order("uuid").
		Offset(offset).
		Limit(limit).
		Find(&userModels)
	if result.Error != nil {
		return nil, fmt.Errorf("get inactive users: %w",
			common.NewDatabaseError("get_inactive_users", "user", "", result.Error))
	}

	users := make([]*entities.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = userModel.ToDomain()
	}

	return users, nil
}

// Search searches users by name or email
func (s *Store) Search(ctx context.Context, query string, offset, limit int) ([]*entities.User, error) {
	var userModels []*UserModel

	result := s.db.GetDB().WithContext(ctx).
		Where("name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%").
		Order("uuid").
		Offset(offset).
		Limit(limit).
		Find(&userModels)
	if result.Error != nil {
		return nil, fmt.Errorf("search users: %w", common.NewDatabaseError("search", "user", query, result.Error))
	}

	users := make([]*entities.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = userModel.ToDomain()
	}

	return users, nil
}
