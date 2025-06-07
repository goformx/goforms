package form

import (
	"context"
	"database/sql"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"strconv"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/jmoiron/sqlx"
)

// FormSubmissionStore implements form.SubmissionRepository
type FormSubmissionStore struct {
	db     *sqlx.DB
	logger logging.Logger
}

// NewFormSubmissionStore creates a new form submission store
func NewFormSubmissionStore(db *sqlx.DB, logger logging.Logger) *FormSubmissionStore {
	return &FormSubmissionStore{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form submission
func (s *FormSubmissionStore) Create(ctx context.Context, submission *model.FormSubmission) error {
	data, err := json.Marshal(submission.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal submission data: %w", err)
	}

	metadata, err := json.Marshal(submission.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal submission metadata: %w", err)
	}

	query := `
		INSERT INTO form_submissions (
			form_id, data, status, submitted_at, metadata, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, NOW(), NOW()
		)
	`

	result, err := s.db.ExecContext(ctx, query,
		submission.FormID,
		data,
		submission.Status,
		submission.SubmittedAt,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to insert submission: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	submission.ID = strconv.FormatInt(id, 10)
	return nil
}

// GetByID retrieves a form submission by its ID
func (s *FormSubmissionStore) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	var submission model.FormSubmission
	query := `SELECT * FROM form_submissions WHERE id = ?`
	err := s.db.GetContext(ctx, &submission, query, id)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.New(domainerrors.ErrCodeNotFound, "form submission not found", nil)
		}
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get form submission")
	}

	return &submission, nil
}

// GetByFormID retrieves all submissions for a specific form
func (s *FormSubmissionStore) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	query := `SELECT * FROM form_submissions WHERE form_id = ? ORDER BY created_at DESC`
	err := s.db.SelectContext(ctx, &submissions, query, formID)
	if err != nil {
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get form submissions")
	}

	return submissions, nil
}

// GetByUserID retrieves all submissions made by a specific user
func (s *FormSubmissionStore) GetByUserID(ctx context.Context, userID uint) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	query := `SELECT * FROM form_submissions WHERE user_id = ? ORDER BY created_at DESC`
	err := s.db.SelectContext(ctx, &submissions, query, userID)
	if err != nil {
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get user submissions")
	}

	return submissions, nil
}

// Update updates a form submission
func (s *FormSubmissionStore) Update(ctx context.Context, submission *model.FormSubmission) error {
	data, err := json.Marshal(submission.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal submission data: %w", err)
	}

	metadata, err := json.Marshal(submission.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal submission metadata: %w", err)
	}

	query := `
		UPDATE form_submissions SET
			data = ?,
			status = ?,
			metadata = ?,
			updated_at = NOW()
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		data,
		submission.Status,
		metadata,
		submission.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update submission: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("submission not found: %s", submission.ID)
	}

	return nil
}

// Delete deletes a form submission
func (s *FormSubmissionStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM form_submissions WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete submission: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("submission not found: %s", id)
	}

	return nil
}

// List retrieves a paginated list of form submissions
func (s *FormSubmissionStore) List(ctx context.Context, offset, limit int) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	query := `SELECT * FROM form_submissions ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := s.db.SelectContext(ctx, &submissions, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list form submissions: %w", err)
	}

	return submissions, nil
}

// Count returns the total number of form submissions
func (s *FormSubmissionStore) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM form_submissions`
	err := s.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count form submissions: %w", err)
	}
	return count, nil
}

// Search searches form submissions by form ID and user ID
func (s *FormSubmissionStore) Search(ctx context.Context, formID string, userID uint, offset, limit int) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	query := `
		SELECT * FROM form_submissions 
		WHERE form_id = ? AND user_id = ? 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`
	err := s.db.SelectContext(ctx, &submissions, query, formID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search form submissions: %w", err)
	}

	return submissions, nil
}
