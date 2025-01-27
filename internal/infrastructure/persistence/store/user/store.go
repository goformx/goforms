package user

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence/database"
)

// Store implements user.Store interface
type Store struct {
	db     *database.DB
	logger logging.Logger
}

// NewStore creates a new user store
func NewStore(db *database.DB, logger logging.Logger) user.Store {
	logger.Debug("creating user store",
		logging.Bool("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create stores a new user
func (s *Store) Create(user *user.User) error {
	query := `
		INSERT INTO users (email, hashed_password, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())
	`

	s.logger.Debug("creating user",
		logging.String("email", user.Email),
	)

	err := s.db.WithTx(context.Background(), func(tx *sqlx.Tx) error {
		result, err := tx.Exec(query,
			user.Email,
			user.HashedPassword,
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

		user.ID = uint(id)
		return nil
	})

	if err != nil {
		s.logger.Error("failed to create user",
			logging.Error(err),
			logging.String("email", user.Email),
		)
		return fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("user created",
		logging.Uint("id", user.ID),
		logging.String("email", user.Email),
	)

	return nil
}

// GetByID returns a user by ID
func (s *Store) GetByID(id uint) (*user.User, error) {
	query := `
		SELECT id, email, hashed_password, created_at, updated_at
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
	)

	return &u, nil
}

// GetByEmail returns a user by email
func (s *Store) GetByEmail(email string) (*user.User, error) {
	query := `
		SELECT id, email, hashed_password, created_at, updated_at
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
	)

	return &u, nil
}

// Update updates user information
func (s *Store) Update(user *user.User) error {
	query := `
		UPDATE users
		SET email = ?, hashed_password = ?, updated_at = NOW()
		WHERE id = ?
	`

	s.logger.Debug("updating user",
		logging.Uint("id", user.ID),
		logging.String("email", user.Email),
	)

	err := s.db.WithTx(context.Background(), func(tx *sqlx.Tx) error {
		result, err := tx.Exec(query,
			user.Email,
			user.HashedPassword,
			user.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rows == 0 {
			return fmt.Errorf("user not found: %d", user.ID)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to update user",
			logging.Error(err),
			logging.Uint("id", user.ID),
			logging.String("email", user.Email),
		)
		return fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("user updated",
		logging.Uint("id", user.ID),
		logging.String("email", user.Email),
	)

	return nil
}

// Delete removes a user
func (s *Store) Delete(id uint) error {
	query := `
		DELETE FROM users
		WHERE id = ?
	`

	s.logger.Debug("deleting user",
		logging.Uint("id", id),
	)

	err := s.db.WithTx(context.Background(), func(tx *sqlx.Tx) error {
		result, err := tx.Exec(query, id)
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

		return nil
	})

	if err != nil {
		s.logger.Error("failed to delete user",
			logging.Error(err),
			logging.Uint("id", id),
		)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.Info("user deleted",
		logging.Uint("id", id),
	)

	return nil
}

// List returns all users
func (s *Store) List() ([]user.User, error) {
	query := `
		SELECT id, email, hashed_password, created_at, updated_at
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
