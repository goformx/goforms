package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/jmoiron/sqlx"
)

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
)

// Store implements user.Store interface
type Store struct {
	db     *sqlx.DB
	logger logging.Logger
}

// NewStore creates a new user store
func NewStore(db *sqlx.DB, logger logging.Logger) user.Store {
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
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO users (email, hashed_password, first_name, last_name, role, active) VALUES (?, ?, ?, ?, ?, ?)",
		u.Email, u.HashedPassword, u.FirstName, u.LastName, u.Role, u.Active,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	if id <= 0 || uint64(id) > uint64(^uint(0)) {
		return fmt.Errorf("user ID %d is out of valid range", id)
	}

	u.ID = uint(id)
	return nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := s.db.GetContext(ctx, &u, "SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &u, nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, active, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var u user.User
	err := s.db.QueryRowxContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.HashedPassword,
		&u.FirstName,
		&u.LastName,
		&u.Role,
		&u.Active,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
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

// Update updates an existing user
func (s *Store) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users 
		SET email = ?, 
			hashed_password = ?, 
			first_name = ?, 
			last_name = ?, 
			role = ?, 
			active = ?, 
			updated_at = NOW() 
		WHERE id = ?
	`
	result, err := s.db.ExecContext(ctx, query,
		u.Email, u.HashedPassword, u.FirstName, u.LastName, u.Role, u.Active, u.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// Delete removes a user by ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// List returns all users
func (s *Store) List(ctx context.Context) ([]user.User, error) {
	var users []user.User
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	if len(users) == 0 {
		return []user.User{}, nil
	}

	return users, nil
}

// ListPaginated returns a paginated list of users
func (s *Store) ListPaginated(ctx context.Context, offset, limit int) ([]*user.User, error) {
	var users []*user.User
	err := s.db.SelectContext(ctx, &users,
		"SELECT * FROM users ORDER BY id LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	if len(users) == 0 {
		return []*user.User{}, nil
	}

	return users, nil
}

// Count returns the total number of users
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users")
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// FindByEmail finds a user by email
func (s *Store) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := s.db.GetContext(ctx, &u, "SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

func (s *Store) FindByID(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	err := s.db.GetContext(ctx, &u, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}
