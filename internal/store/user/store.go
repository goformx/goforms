package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
)

// Store implements the UserStore interface using a SQL database
type Store struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewStore creates a new user store
func NewStore(db *sqlx.DB, log logger.Logger) models.UserStore {
	return &Store{
		db:  db,
		log: log,
	}
}

// Create inserts a new user into the database
func (s *Store) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, hashed_password, first_name, last_name, role, active, created_at, updated_at)
		VALUES (:email, :hashed_password, :first_name, :last_name, :role, :active, :created_at, :updated_at)
		RETURNING id`

	rows, err := s.db.NamedQuery(query, user)
	if err != nil {
		s.log.Error("failed to create user", logger.Error(err))
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&user.ID)
		if err != nil {
			s.log.Error("failed to scan user id", logger.Error(err))
			return err
		}
	}

	return nil
}

// GetByID retrieves a user by their ID
func (s *Store) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		s.log.Error("failed to get user by id", logger.Error(err))
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by their email
func (s *Store) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		s.log.Error("failed to get user by email", logger.Error(err))
		return nil, err
	}
	return &user, nil
}

// Update updates an existing user
func (s *Store) Update(user *models.User) error {
	user.UpdatedAt = time.Now()
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

	result, err := s.db.NamedExec(query, user)
	if err != nil {
		s.log.Error("failed to update user", logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		s.log.Error("failed to get rows affected", logger.Error(err))
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete removes a user from the database
func (s *Store) Delete(id uint) error {
	result, err := s.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		s.log.Error("failed to delete user", logger.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		s.log.Error("failed to get rows affected", logger.Error(err))
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// List returns all users
func (s *Store) List() ([]models.User, error) {
	var users []models.User
	err := s.db.Select(&users, "SELECT * FROM users")
	if err != nil {
		s.log.Error("failed to list users", logger.Error(err))
		return nil, err
	}
	return users, nil
}
