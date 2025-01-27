package contact

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence/database"
)

// Store implements contact.Store interface
type Store struct {
	db     *database.DB
	logger logging.Logger
}

// NewStore creates a new contact store
func NewStore(db *database.DB, logger logging.Logger) contact.Store {
	logger.Debug("creating contact store",
		logging.Bool("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create stores a new contact form submission
func (s *Store) Create(ctx context.Context, sub *contact.Submission) error {
	query := `
		INSERT INTO contact_submissions (name, email, message, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	s.logger.Debug("creating contact submission",
		logging.String("email", sub.Email),
		logging.String("status", string(sub.Status)),
	)

	err := s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, query,
			sub.Name,
			sub.Email,
			sub.Message,
			sub.Status,
			sub.CreatedAt,
			sub.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert contact submission: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert ID: %w", err)
		}

		sub.ID = id
		return nil
	})

	if err != nil {
		s.logger.Error("failed to create contact submission",
			logging.Error(err),
			logging.String("email", sub.Email),
		)
		return fmt.Errorf("failed to create contact submission: %w", err)
	}

	s.logger.Info("contact submission created",
		logging.Int64("id", sub.ID),
		logging.String("email", sub.Email),
		logging.String("status", string(sub.Status)),
	)

	return nil
}

// List returns all contact form submissions
func (s *Store) List(ctx context.Context) ([]contact.Submission, error) {
	query := `
		SELECT id, name, email, message, status, created_at, updated_at
		FROM contact_submissions
		ORDER BY created_at DESC
	`

	s.logger.Debug("listing contact submissions")

	var submissions []contact.Submission
	if err := s.db.SelectContext(ctx, &submissions, query); err != nil {
		s.logger.Error("failed to list contact submissions",
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to list contact submissions: %w", err)
	}

	s.logger.Debug("contact submissions retrieved",
		logging.Int("count", len(submissions)),
	)

	return submissions, nil
}

// Get returns a specific contact form submission
func (s *Store) Get(ctx context.Context, id int64) (*contact.Submission, error) {
	query := `
		SELECT id, name, email, message, status, created_at, updated_at
		FROM contact_submissions
		WHERE id = ?
	`

	s.logger.Debug("getting contact submission",
		logging.Int64("id", id),
	)

	var submission contact.Submission
	if err := s.db.GetContext(ctx, &submission, query, id); err != nil {
		s.logger.Error("failed to get contact submission",
			logging.Error(err),
			logging.Int64("id", id),
		)
		return nil, fmt.Errorf("failed to get contact submission: %w", err)
	}

	s.logger.Debug("contact submission retrieved",
		logging.Int64("id", submission.ID),
		logging.String("email", submission.Email),
		logging.String("status", string(submission.Status)),
	)

	return &submission, nil
}

// UpdateStatus updates the status of a contact form submission
func (s *Store) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	query := `
		UPDATE contact_submissions
		SET status = ?, updated_at = NOW()
		WHERE id = ?
	`

	s.logger.Debug("updating contact submission status",
		logging.Int64("id", id),
		logging.String("status", string(status)),
	)

	err := s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, query, status, id)
		if err != nil {
			return fmt.Errorf("failed to update contact submission status: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rows == 0 {
			return fmt.Errorf("contact submission not found: %d", id)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to update contact submission status",
			logging.Error(err),
			logging.Int64("id", id),
			logging.String("status", string(status)),
		)
		return fmt.Errorf("failed to update contact submission status: %w", err)
	}

	s.logger.Info("contact submission status updated",
		logging.Int64("id", id),
		logging.String("status", string(status)),
	)

	return nil
}
