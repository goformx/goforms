package store

import (
	"database/sql"
	"errors"
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

// NewStore creates a new user store
func NewStore(db *sqlx.DB, log logging.Logger) user.Store {
	return &Store{
		db:  db,
		log: log,
	}
}

// Create inserts a new user into the database
func (s *Store) Create(u *user.User) error {
	query := `
		INSERT INTO users (email, hashed_password, first_name, last_name, role, active, created_at, updated_at)
		VALUES (:email, :hashed_password, :first_name, :last_name, :role, :active, :created_at, :updated_at)
		RETURNING id`

	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	rows, err := s.db.NamedQuery(query, u)
	if err != nil {
		s.log.Error("failed to create user", logging.Error(err))
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&u.ID); err != nil {
			s.log.Error("failed to scan user id", logging.Error(err))
			return err
		}
	}

	return nil
}

// GetByID retrieves a user by their ID
func (s *Store) GetByID(id uint) (*user.User, error) {
	var u user.User
	err := s.db.Get(&u, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		s.log.Error("failed to get user by id", logging.Error(err))
		return nil, err
	}
	return &u, nil
}

// GetByEmail retrieves a user by their email address
func (s *Store) GetByEmail(email string) (*user.User, error) {
	var u user.User
	err := s.db.Get(&u, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		s.log.Error("failed to get user by email", logging.Error(err))
		return nil, err
	}
	return &u, nil
}

// Update modifies an existing user in the database
func (s *Store) Update(u *user.User) error {
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

	result, err := s.db.NamedExec(query, u)
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
func (s *Store) Delete(id uint) error {
	result, err := s.db.Exec("DELETE FROM users WHERE id = $1", id)
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
func (s *Store) List() ([]user.User, error) {
	var users []user.User
	err := s.db.Select(&users, "SELECT * FROM users")
	if err != nil {
		s.log.Error("failed to list users", logging.Error(err))
		return nil, err
	}
	return users, nil
}
