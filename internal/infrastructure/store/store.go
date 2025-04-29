package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Store implements the user.Store interface using a SQL database
type Store struct {
	db  *sqlx.DB
	log logging.Logger
}

var (
	ErrUserNotFound = errors.New("user not found")
)

// NewStore creates a new user store
func NewStore(db *sqlx.DB, log logging.Logger) user.Store {
	return &Store{
		db:  db,
		log: log,
	}
}

// Create inserts a new user into the database
func (s *Store) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (email, hashed_password, first_name, last_name, role, active, created_at, updated_at)
		VALUES (:email, :hashed_password, :first_name, :last_name, :role, :active, :created_at, :updated_at)
		RETURNING id`

	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	rows, queryErr := s.db.NamedQueryContext(ctx, query, u)
	if queryErr != nil {
		s.log.Error("failed to create user", logging.Error(queryErr))
		return queryErr
	}
	defer rows.Close()

	if rows.Next() {
		if scanErr := rows.Scan(&u.ID); scanErr != nil {
			s.log.Error("failed to scan user id", logging.Error(scanErr))
			return scanErr
		}
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return rowsErr
	}

	return nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*user.User, error) {
	query := `SELECT * FROM users WHERE id = $1`
	var u user.User
	err := s.db.GetContext(ctx, &u, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.log.Error("failed to get user by ID", logging.Error(err))
		return nil, err
	}
	return &u, nil
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT * FROM users WHERE email = $1`
	var u user.User
	err := s.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.log.Error("failed to get user by email", logging.Error(err))
		return nil, err
	}
	return &u, nil
}

// Update modifies an existing user in the database
func (s *Store) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET email = :email,
			hashed_password = :hashed_password,
			first_name = :first_name,
			last_name = :last_name,
			role = :role,
			active = :active,
			updated_at = :updated_at
		WHERE id = :id`

	u.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, u)
	if err != nil {
		s.log.Error("failed to update user", logging.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.log.Error("failed to get rows affected", logging.Error(err))
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete removes a user from the database
func (s *Store) Delete(ctx context.Context, id uint) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		s.log.Error("failed to delete user", logging.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.log.Error("failed to get rows affected", logging.Error(err))
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// List returns all users from the database
func (s *Store) List(ctx context.Context) ([]user.User, error) {
	var users []user.User
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM users")
	if err != nil {
		s.log.Error("failed to list users", logging.Error(err))
		return nil, err
	}
	return users, nil
}

func (s *Store) GetUserIDs(ctx context.Context) ([]string, error) {
	query := "SELECT id FROM users"
	rows, queryErr := s.db.QueryContext(ctx, query)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to query user IDs: %w", queryErr)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var u struct {
			ID string
		}
		if scanErr := rows.Scan(&u.ID); scanErr != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", scanErr)
		}
		userIDs = append(userIDs, u.ID)
	}

	if rowErr := rows.Err(); rowErr != nil {
		return nil, fmt.Errorf("error iterating rows: %w", rowErr)
	}

	return userIDs, nil
}
