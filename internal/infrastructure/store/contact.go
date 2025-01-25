package store

import (
	"context"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

type contactStore struct {
	db     *database.Database
	logger logging.Logger
}

// NewContactStore creates a new contact store
func NewContactStore(db *database.Database, logger logging.Logger) contact.Store {
	return &contactStore{
		db:     db,
		logger: logger,
	}
}

// Create implements contact.Store
func (s *contactStore) Create(ctx context.Context, submission *contact.Submission) error {
	query := `INSERT INTO contact_submissions (name, email, message, status, created_at, updated_at)
			  VALUES (:name, :email, :message, :status, NOW(), NOW())`

	_, err := s.db.NamedExecContext(ctx, query, submission)
	if err != nil {
		s.logger.Error("failed to create contact submission", logging.Error(err))
		return err
	}

	return nil
}

// List implements contact.Store
func (s *contactStore) List(ctx context.Context) ([]contact.Submission, error) {
	var submissions []contact.Submission
	query := `SELECT * FROM contact_submissions ORDER BY created_at DESC`

	if err := s.db.SelectContext(ctx, &submissions, query); err != nil {
		s.logger.Error("failed to list contact submissions", logging.Error(err))
		return nil, err
	}

	return submissions, nil
}

// Get implements contact.Store
func (s *contactStore) Get(ctx context.Context, id int64) (*contact.Submission, error) {
	var submission contact.Submission
	query := `SELECT * FROM contact_submissions WHERE id = ?`

	if err := s.db.GetContext(ctx, &submission, query, id); err != nil {
		s.logger.Error("failed to get contact submission", logging.Error(err))
		return nil, err
	}

	return &submission, nil
}

// UpdateStatus implements contact.Store
func (s *contactStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	query := `UPDATE contact_submissions SET status = ?, updated_at = NOW() WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		s.logger.Error("failed to update contact submission status", logging.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("failed to get rows affected", logging.Error(err))
		return err
	}

	if rows == 0 {
		return contact.ErrSubmissionNotFound
	}

	return nil
}

// Other required methods for contact.Store interface...
