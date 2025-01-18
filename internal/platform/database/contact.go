package database

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/jonesrussell/goforms/internal/logger"
)

type ContactStore struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewContactStore(db *sqlx.DB, logger logger.Logger) contact.Store {
	return &ContactStore{
		db:     db,
		logger: logger,
	}
}

func (s *ContactStore) Create(ctx context.Context, submission *contact.Submission) error {
	query := `
        INSERT INTO contact_submissions (name, email, message, status)
        VALUES (:name, :email, :message, :status)
    `
	_, err := s.db.NamedExecContext(ctx, query, submission)
	if err != nil {
		return err
	}
	return nil
}

func (s *ContactStore) List(ctx context.Context) ([]contact.Submission, error) {
	var submissions []contact.Submission
	query := `
        SELECT id, name, email, message, status, created_at, updated_at
        FROM contact_submissions 
        ORDER BY created_at DESC
    `
	if err := s.db.SelectContext(ctx, &submissions, query); err != nil {
		return nil, err
	}
	return submissions, nil
}

func (s *ContactStore) GetByID(ctx context.Context, id int64) (*contact.Submission, error) {
	var submission contact.Submission
	query := `
        SELECT id, name, email, message, status, created_at, updated_at
        FROM contact_submissions 
        WHERE id = ?
    `
	if err := s.db.GetContext(ctx, &submission, query, id); err != nil {
		return nil, err
	}
	return &submission, nil
}

func (s *ContactStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	query := `
        UPDATE contact_submissions 
        SET status = :status 
        WHERE id = :id
    `
	_, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":     id,
		"status": status,
	})
	return err
}
