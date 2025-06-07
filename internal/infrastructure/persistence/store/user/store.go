package user

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
)

// Store implements user.Repository interface
type Store struct {
	db     *database.Database
	logger logging.Logger
}

// NewStore creates a new user store
func NewStore(db *database.Database, logger logging.Logger) user.Repository {
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
	query := `
		INSERT INTO users (
			email, hashed_password, first_name, last_name, role, active, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, NOW(), NOW()
		) RETURNING id
	`

	var id uint
	err := s.db.GetContext(ctx, &id, query,
		u.Email, u.HashedPassword, u.FirstName, u.LastName, u.Role, u.Active,
	)
	if err != nil {
		return domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to insert user")
	}

	u.ID = id
	return nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	query := `SELECT * FROM users WHERE email = ?`
	err := s.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
		}
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get user")
	}

	// Convert timestamps to UTC
	u.CreatedAt = u.CreatedAt.UTC()
	u.UpdatedAt = u.UpdatedAt.UTC()

	return &u, nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	query := `SELECT * FROM users WHERE id = ?`
	err := s.db.GetContext(ctx, &u, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
		}
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get user")
	}

	// Convert timestamps to UTC
	u.CreatedAt = u.CreatedAt.UTC()
	u.UpdatedAt = u.UpdatedAt.UTC()

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
	query := `
		UPDATE users 
		SET email = ?, hashed_password = ?, first_name = ?, last_name = ?, role = ?, active = ?, updated_at = NOW()
		WHERE id = ?
	`
	result, err := s.db.ExecContext(ctx, query,
		userModel.Email,
		userModel.HashedPassword,
		userModel.FirstName,
		userModel.LastName,
		userModel.Role,
		userModel.Active,
		userModel.ID,
	)
	if err != nil {
		return domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to update user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get rows affected")
	}

	if rows == 0 {
		return domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
	}

	return nil
}

// Delete removes a user by ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	query := `DELETE FROM users WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to delete user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get rows affected")
	}

	if rows == 0 {
		return domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
	}

	return nil
}

// List returns all users
func (s *Store) List(ctx context.Context) ([]user.User, error) {
	var users []user.User
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM users ORDER BY id")
	if err != nil {
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to list users")
	}

	if len(users) == 0 {
		return []user.User{}, nil
	}

	return users, nil
}

// ListPaginated returns a paginated list of users
func (s *Store) ListPaginated(ctx context.Context, offset, limit int) ([]*user.User, error) {
	query := `SELECT * FROM users ORDER BY id LIMIT ? OFFSET ?`
	var users []*user.User
	err := s.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to list users")
	}
	return users, nil
}

// Count returns the total number of users
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users")
	if err != nil {
		return 0, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to count users")
	}
	return count, nil
}
