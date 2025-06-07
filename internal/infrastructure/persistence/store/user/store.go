package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "create_user"),
				logging.Error(rollbackErr),
			)
		}
	}()

	query := `
		INSERT INTO users (
			email, hashed_password, first_name, last_name, role, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		) RETURNING id
	`

	var id uint
	err = tx.GetContext(ctx, &id, query,
		u.Email, u.HashedPassword, u.FirstName, u.LastName, u.Role, u.Active,
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	u.ID = id

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	query := fmt.Sprintf(`SELECT * FROM users WHERE email = %s`, s.db.GetPlaceholder(1))
	err := s.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.WrapError(err, domainerrors.ErrCodeNotFound, "user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	query := fmt.Sprintf(`SELECT * FROM users WHERE id = %s`, s.db.GetPlaceholder(1))
	err := s.db.GetContext(ctx, &u, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
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
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "update_user"),
				logging.Error(rollbackErr),
			)
		}
	}()

	query := fmt.Sprintf(`
		UPDATE users SET
			email = %s,
			hashed_password = %s,
			first_name = %s,
			last_name = %s,
			role = %s,
			active = %s,
			updated_at = NOW()
		WHERE id = %s
	`,
		s.db.GetPlaceholder(1),
		s.db.GetPlaceholder(2),
		s.db.GetPlaceholder(3),
		s.db.GetPlaceholder(4),
		s.db.GetPlaceholder(5),
		s.db.GetPlaceholder(6),
		s.db.GetPlaceholder(7),
	)

	result, err := tx.ExecContext(ctx, query,
		userModel.Email,
		userModel.HashedPassword,
		userModel.FirstName,
		userModel.LastName,
		userModel.Role,
		userModel.Active,
		userModel.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %d", userModel.ID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete removes a user by ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "delete_user"),
				logging.Error(rollbackErr),
			)
		}
	}()

	query := fmt.Sprintf(`DELETE FROM users WHERE id = %s`, s.db.GetPlaceholder(1))
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %d", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
	query := fmt.Sprintf(`SELECT * FROM users ORDER BY id LIMIT %s OFFSET %s`,
		s.db.GetPlaceholder(1),
		s.db.GetPlaceholder(2),
	)
	var users []*user.User
	err := s.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
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
	query := fmt.Sprintf(`SELECT * FROM users WHERE email = %s`, s.db.GetPlaceholder(1))
	var u user.User
	err := s.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.WrapError(err, domainerrors.ErrCodeNotFound, "user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}

// FindByID finds a user by ID
func (s *Store) FindByID(ctx context.Context, id uint) (*user.User, error) {
	query := fmt.Sprintf(`SELECT * FROM users WHERE id = %s`, s.db.GetPlaceholder(1))
	var u user.User
	err := s.db.GetContext(ctx, &u, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}
