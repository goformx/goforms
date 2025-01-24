package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// ContactStore implements contact.Store
type ContactStore struct {
	db     *sqlx.DB
	logger logging.Logger
}

// NewContactStore creates a new contact store
func NewContactStore(db *DB, logger logging.Logger) contact.Store {
	return &ContactStore{
		db:     db.DB,
		logger: logger,
	}
}

// Create creates a new contact submission
func (s *ContactStore) Create(ctx context.Context, sub *contact.Submission) error {
	query := `
		INSERT INTO contact_submissions (name, email, message, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		sub.Name,
		sub.Email,
		sub.Message,
		sub.Status,
		sub.CreatedAt,
		sub.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	sub.ID = id
	return nil
}

// List returns all contact submissions
func (s *ContactStore) List(ctx context.Context) ([]contact.Submission, error) {
	var subs []contact.Submission
	query := `
		SELECT id, name, email, message, status, created_at, updated_at
		FROM contact_submissions
		ORDER BY created_at DESC
	`

	if err := s.db.SelectContext(ctx, &subs, query); err != nil {
		return nil, err
	}

	return subs, nil
}

// Get returns a contact submission by ID
func (s *ContactStore) Get(ctx context.Context, id int64) (*contact.Submission, error) {
	var sub contact.Submission
	query := `
		SELECT id, name, email, message, status, created_at, updated_at
		FROM contact_submissions
		WHERE id = ?
	`

	err := s.db.GetContext(ctx, &sub, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &sub, nil
}

// UpdateStatus updates the status of a contact submission
func (s *ContactStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	query := `
		UPDATE contact_submissions
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("contact submission not found")
	}

	return nil
}
