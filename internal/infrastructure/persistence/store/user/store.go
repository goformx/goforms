package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
)

// Store implements user.Store interface
type Store struct {
	db     *database.Database
	logger logging.Logger
}

// NewStore creates a new user store
func NewStore(db *database.Database, logger logging.Logger) user.Store {
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
		INSERT INTO users (email, hashed_password, first_name, last_name, role, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	s.logger.Debug("creating user",
		logging.StringField("email", u.Email),
		logging.StringField("role", u.Role),
		logging.BoolField("active", u.Active),
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Error("failed to rollback transaction",
					logging.ErrorField("error", rbErr),
				)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			s.logger.Error("failed to commit transaction",
				logging.ErrorField("error", err),
			)
		}
	}()

	result, err := tx.ExecContext(ctx, query,
		u.Email,
		u.HashedPassword,
		u.FirstName,
		u.LastName,
		u.Role,
		u.Active,
		u.CreatedAt,
		u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	if id <= 0 || uint64(id) > uint64(^uint(0)) {
		return fmt.Errorf("user ID %d is out of valid range", id)
	}

	u.ID = uint(id)

	s.logger.Info("user created",
		logging.UintField("id", u.ID),
		logging.StringField("email", u.Email),
	)

	return nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, active, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	s.logger.Debug("getting user by email",
		logging.StringField("email", email),
	)

	var u user.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
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

	if err == sql.ErrNoRows {
		s.logger.Debug("user not found by email",
			logging.StringField("email", email),
		)
		return nil, ErrUserNotFound
	}
	if err != nil {
		s.logger.Error("failed to get user by email",
			logging.ErrorField("error", err),
			logging.StringField("email", email),
		)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	s.logger.Debug("user found by email",
		logging.UintField("id", u.ID),
		logging.StringField("email", u.Email),
	)
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
	err := s.db.QueryRowContext(ctx, query, id).Scan(
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

	if err == sql.ErrNoRows {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &u, nil
}

// Update updates an existing user
func (s *Store) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET email = ?, hashed_password = ?, first_name = ?, last_name = ?, role = ?, active = ?, updated_at = ?
		WHERE id = ?
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Error("failed to rollback transaction",
					logging.ErrorField("error", rbErr),
				)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			s.logger.Error("failed to commit transaction",
				logging.ErrorField("error", err),
			)
		}
	}()

	result, err := tx.ExecContext(ctx, query,
		u.Email,
		u.HashedPassword,
		u.FirstName,
		u.LastName,
		u.Role,
		u.Active,
		time.Now(),
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	s.logger.Info("user updated",
		logging.UintField("id", u.ID),
		logging.StringField("email", u.Email),
	)

	return nil
}

// Delete removes a user by ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	query := `DELETE FROM users WHERE id = ?`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Error("failed to rollback transaction",
					logging.ErrorField("error", rbErr),
				)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			s.logger.Error("failed to commit transaction",
				logging.ErrorField("error", err),
			)
		}
	}()

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	s.logger.Info("user deleted",
		logging.UintField("id", id),
	)

	return nil
}

// List returns all users
func (s *Store) List(ctx context.Context) ([]user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, active, created_at, updated_at
		FROM users
		ORDER BY id
	`

	rows, queryErr := s.db.QueryContext(ctx, query)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to list users: %w", queryErr)
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var u user.User
		scanErr := rows.Scan(
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
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan user: %w", scanErr)
		}
		users = append(users, u)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating users: %w", rowsErr)
	}

	return users, nil
}
