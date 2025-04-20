package user

import (
	"context"
	"fmt"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Store implements user.Store interface
type Store struct {
	db     *database.Database
	logger logging.Logger
}

// NewStore creates a new user store
func NewStore(db *database.Database, logger logging.Logger) user.Store {
	logger.Debug("creating user store",
		logging.Bool("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create stores a new user
func (s *Store) Create(u *user.User) error {
	query := `
		INSERT INTO users (email, hashed_password, first_name, last_name, role, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	s.logger.Debug("creating user",
		logging.String("email", u.Email),
		logging.String("role", u.Role),
		logging.Bool("active", u.Active),
	)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Error("failed to rollback transaction",
					logging.Error(rbErr),
				)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			s.logger.Error("failed to commit transaction",
				logging.Error(err),
			)
		}
	}()

	result, err := tx.ExecContext(context.Background(), query,
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

	// Check for integer overflow before conversion
	if id <= 0 || uint64(id) > uint64(^uint(0)) {
		return fmt.Errorf("user ID %d is out of valid range", id)
	}

	u.ID = uint(id)

	if err != nil {
		s.logger.Error("failed to create user",
			logging.Error(err),
			logging.String("email", u.Email),
		)
		return fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("user created",
		logging.Uint("id", u.ID),
		logging.String("email", u.Email),
		logging.String("role", u.Role),
		logging.Bool("active", u.Active),
	)

	return nil
}

// GetByEmail returns a user by email
func (s *Store) GetByEmail(email string) (*user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, active, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	s.logger.Debug("getting user by email",
		logging.String("email", email),
	)

	var u user.User
	if err := s.db.Get(&u, query, email); err != nil {
		s.logger.Error("failed to get user by email",
			logging.Error(err),
			logging.String("email", email),
		)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	s.logger.Debug("user retrieved",
		logging.Uint("id", u.ID),
		logging.String("email", u.Email),
		logging.String("role", u.Role),
		logging.Bool("active", u.Active),
	)

	return &u, nil
}

// GetByID returns a user by ID
func (s *Store) GetByID(id uint) (*user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, active, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	s.logger.Debug("getting user by ID",
		logging.Uint("id", id),
	)

	var u user.User
	if err := s.db.Get(&u, query, id); err != nil {
		s.logger.Error("failed to get user by ID",
			logging.Error(err),
			logging.Uint("id", id),
		)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	s.logger.Debug("user retrieved",
		logging.Uint("id", u.ID),
		logging.String("email", u.Email),
		logging.String("role", u.Role),
		logging.Bool("active", u.Active),
	)

	return &u, nil
}

// Update updates a user
func (s *Store) Update(u *user.User) error {
	query := `
		UPDATE users
		SET email = ?, hashed_password = ?, first_name = ?, last_name = ?, role = ?, active = ?, updated_at = ?
		WHERE id = ?
	`

	s.logger.Debug("updating user",
		logging.Uint("id", u.ID),
		logging.String("email", u.Email),
		logging.String("role", u.Role),
		logging.Bool("active", u.Active),
	)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Error("failed to rollback transaction",
					logging.Error(rbErr),
				)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			s.logger.Error("failed to commit transaction",
				logging.Error(err),
			)
		}
	}()

	result, err := tx.ExecContext(context.Background(), query,
		u.Email,
		u.HashedPassword,
		u.FirstName,
		u.LastName,
		u.Role,
		u.Active,
		u.UpdatedAt,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %d", u.ID)
	}

	s.logger.Info("user updated",
		logging.Uint("id", u.ID),
		logging.String("email", u.Email),
		logging.String("role", u.Role),
		logging.Bool("active", u.Active),
	)

	return nil
}

// Delete deletes a user
func (s *Store) Delete(id uint) error {
	query := `
		DELETE FROM users
		WHERE id = ?
	`

	s.logger.Debug("deleting user",
		logging.Uint("id", id),
	)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Error("failed to rollback transaction",
					logging.Error(rbErr),
				)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			s.logger.Error("failed to commit transaction",
				logging.Error(err),
			)
		}
	}()

	result, err := tx.ExecContext(context.Background(), query, id)
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

	s.logger.Info("user deleted",
		logging.Uint("id", id),
	)

	return nil
}

// List returns all users
func (s *Store) List() ([]user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	s.logger.Debug("listing users")

	var users []user.User
	if err := s.db.Select(&users, query); err != nil {
		s.logger.Error("failed to list users",
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	s.logger.Debug("users retrieved",
		logging.Int("count", len(users)),
	)

	return users, nil
}
