package form

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "create_submission"),
				logging.Error(err),
			)
		}
	}()

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

	result, err := tx.ExecContext(ctx, query,
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a form submission by ID
func (s *FormSubmissionStore) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	var submission struct {
		ID          uint      `db:"id"`
		FormID      string    `db:"form_id"`
		Data        string    `db:"data"`
		Status      string    `db:"status"`
		SubmittedAt time.Time `db:"submitted_at"`
		Metadata    string    `db:"metadata"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	query := `SELECT * FROM form_submissions WHERE id = ?`
	err := s.db.GetContext(ctx, &submission, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("submission not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	var data model.JSON
	if err := json.Unmarshal([]byte(submission.Data), &data); err != nil {
		s.logger.Error("failed to unmarshal submission data",
			logging.String("operation", "get_submission"),
			logging.Error(err),
		)
		data = model.JSON{}
	}

	var metadata model.JSON
	if err := json.Unmarshal([]byte(submission.Metadata), &metadata); err != nil {
		s.logger.Error("failed to unmarshal submission metadata",
			logging.String("operation", "get_submission"),
			logging.Error(err),
		)
		metadata = model.JSON{}
	}

	return &model.FormSubmission{
		ID:          strconv.FormatUint(uint64(submission.ID), 10),
		FormID:      submission.FormID,
		Data:        data,
		SubmittedAt: submission.SubmittedAt,
		Status:      model.SubmissionStatus(submission.Status),
		Metadata:    metadata,
	}, nil
}

// GetByFormID retrieves all submissions for a form
func (s *FormSubmissionStore) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	var submissions []struct {
		ID          uint      `db:"id"`
		FormID      string    `db:"form_id"`
		Data        string    `db:"data"`
		Status      string    `db:"status"`
		SubmittedAt time.Time `db:"submitted_at"`
		Metadata    string    `db:"metadata"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	query := `SELECT * FROM form_submissions WHERE form_id = ? ORDER BY created_at DESC`
	err := s.db.SelectContext(ctx, &submissions, query, formID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	result := make([]*model.FormSubmission, len(submissions))
	for i, submission := range submissions {
		var data model.JSON
		if err := json.Unmarshal([]byte(submission.Data), &data); err != nil {
			s.logger.Error("failed to unmarshal submission data",
				logging.String("operation", "get_submissions"),
				logging.Error(err),
			)
			data = model.JSON{}
		}

		var metadata model.JSON
		if err := json.Unmarshal([]byte(submission.Metadata), &metadata); err != nil {
			s.logger.Error("failed to unmarshal submission metadata",
				logging.String("operation", "get_submissions"),
				logging.Error(err),
			)
			metadata = model.JSON{}
		}

		result[i] = &model.FormSubmission{
			ID:          strconv.FormatUint(uint64(submission.ID), 10),
			FormID:      submission.FormID,
			Data:        data,
			SubmittedAt: submission.SubmittedAt,
			Status:      model.SubmissionStatus(submission.Status),
			Metadata:    metadata,
		}
	}

	return result, nil
}

// GetByUserID retrieves all submissions by a user
func (s *FormSubmissionStore) GetByUserID(ctx context.Context, userID uint) ([]*model.FormSubmission, error) {
	var submissions []struct {
		ID          uint      `db:"id"`
		FormID      string    `db:"form_id"`
		Data        string    `db:"data"`
		Status      string    `db:"status"`
		SubmittedAt time.Time `db:"submitted_at"`
		Metadata    string    `db:"metadata"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	query := `SELECT * FROM form_submissions WHERE user_id = ?`
	err := s.db.SelectContext(ctx, &submissions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	result := make([]*model.FormSubmission, len(submissions))
	for i, submission := range submissions {
		var data model.JSON
		if err := json.Unmarshal([]byte(submission.Data), &data); err != nil {
			s.logger.Error("failed to unmarshal submission data",
				logging.String("operation", "get_submissions"),
				logging.Error(err),
			)
			data = model.JSON{}
		}

		var metadata model.JSON
		if err := json.Unmarshal([]byte(submission.Metadata), &metadata); err != nil {
			s.logger.Error("failed to unmarshal submission metadata",
				logging.String("operation", "get_submissions"),
				logging.Error(err),
			)
			metadata = model.JSON{}
		}

		result[i] = &model.FormSubmission{
			ID:          strconv.FormatUint(uint64(submission.ID), 10),
			FormID:      submission.FormID,
			Data:        data,
			SubmittedAt: submission.SubmittedAt,
			Status:      model.SubmissionStatus(submission.Status),
			Metadata:    metadata,
		}
	}

	return result, nil
}

// Update updates a form submission
func (s *FormSubmissionStore) Update(ctx context.Context, submission *model.FormSubmission) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "update_submission"),
				logging.Error(err),
			)
		}
	}()

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
			submitted_at = ?,
			metadata = ?,
			updated_at = NOW()
		WHERE id = ?
	`

	result, err := tx.ExecContext(ctx, query,
		data,
		submission.Status,
		submission.SubmittedAt,
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete deletes a form submission
func (s *FormSubmissionStore) Delete(ctx context.Context, id string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "delete_submission"),
				logging.Error(err),
			)
		}
	}()

	query := `DELETE FROM form_submissions WHERE id = ?`

	result, err := tx.ExecContext(ctx, query, id)
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// List returns a paginated list of form submissions
func (s *FormSubmissionStore) List(ctx context.Context, offset, limit int) ([]*model.FormSubmission, error) {
	var submissions []struct {
		ID          uint      `db:"id"`
		FormID      string    `db:"form_id"`
		Data        string    `db:"data"`
		Status      string    `db:"status"`
		SubmittedAt time.Time `db:"submitted_at"`
		Metadata    string    `db:"metadata"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	query := `SELECT * FROM form_submissions ORDER BY submitted_at DESC LIMIT ? OFFSET ?`
	err := s.db.SelectContext(ctx, &submissions, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}

	result := make([]*model.FormSubmission, len(submissions))
	for i, submission := range submissions {
		var data model.JSON
		if err := json.Unmarshal([]byte(submission.Data), &data); err != nil {
			s.logger.Error("failed to unmarshal submission data",
				logging.String("operation", "list_submissions"),
				logging.Error(err),
			)
			data = model.JSON{}
		}

		var metadata model.JSON
		if err := json.Unmarshal([]byte(submission.Metadata), &metadata); err != nil {
			s.logger.Error("failed to unmarshal submission metadata",
				logging.String("operation", "list_submissions"),
				logging.Error(err),
			)
			metadata = model.JSON{}
		}

		result[i] = &model.FormSubmission{
			ID:          strconv.FormatUint(uint64(submission.ID), 10),
			FormID:      submission.FormID,
			Data:        data,
			SubmittedAt: submission.SubmittedAt,
			Status:      model.SubmissionStatus(submission.Status),
			Metadata:    metadata,
		}
	}

	return result, nil
}

// Count returns the total number of form submissions
func (s *FormSubmissionStore) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM form_submissions")
	if err != nil {
		return 0, fmt.Errorf("failed to count submissions: %w", err)
	}
	return count, nil
}

// Search searches form submissions by form ID or user ID
func (s *FormSubmissionStore) Search(ctx context.Context, formID string, userID uint, offset, limit int) ([]*model.FormSubmission, error) {
	var submissions []struct {
		ID          uint      `db:"id"`
		FormID      string    `db:"form_id"`
		Data        string    `db:"data"`
		Status      string    `db:"status"`
		SubmittedAt time.Time `db:"submitted_at"`
		Metadata    string    `db:"metadata"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	query := `
		SELECT * FROM form_submissions 
		WHERE form_id = ? AND user_id = ?
		ORDER BY submitted_at DESC 
		LIMIT ? OFFSET ?
	`
	err := s.db.SelectContext(ctx, &submissions, query, formID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search submissions: %w", err)
	}

	result := make([]*model.FormSubmission, len(submissions))
	for i, submission := range submissions {
		var data model.JSON
		if err := json.Unmarshal([]byte(submission.Data), &data); err != nil {
			s.logger.Error("failed to unmarshal submission data",
				logging.String("operation", "search_submissions"),
				logging.Error(err),
			)
			data = model.JSON{}
		}

		var metadata model.JSON
		if err := json.Unmarshal([]byte(submission.Metadata), &metadata); err != nil {
			s.logger.Error("failed to unmarshal submission metadata",
				logging.String("operation", "search_submissions"),
				logging.Error(err),
			)
			metadata = model.JSON{}
		}

		result[i] = &model.FormSubmission{
			ID:          strconv.FormatUint(uint64(submission.ID), 10),
			FormID:      submission.FormID,
			Data:        data,
			SubmittedAt: submission.SubmittedAt,
			Status:      model.SubmissionStatus(submission.Status),
			Metadata:    metadata,
		}
	}

	return result, nil
}
