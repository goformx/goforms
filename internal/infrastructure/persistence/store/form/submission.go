package form

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// formSubmissionStore implements form.SubmissionStore interface
type formSubmissionStore struct {
	db     *database.Database
	logger logging.Logger
}

// NewSubmissionStore creates a new form submission store
func NewSubmissionStore(db *database.Database, logger logging.Logger) form.SubmissionStore {
	logger.Debug("creating form submission store",
		logging.BoolField("db_available", db != nil),
	)
	return &formSubmissionStore{
		db:     db,
		logger: logger,
	}
}

func (s *formSubmissionStore) Create(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("Create called", logging.StringField("form_id", submission.FormID))
	query := fmt.Sprintf(`INSERT INTO form_submissions (uuid, form_uuid, data, submitted_at, status, metadata) 
		VALUES (%s, %s, %s, %s, %s, %s)`,
		s.db.GetPlaceholder(1),
		s.db.GetPlaceholder(2),
		s.db.GetPlaceholder(3),
		s.db.GetPlaceholder(4),
		s.db.GetPlaceholder(5),
		s.db.GetPlaceholder(6),
	)

	dataBytes, err := json.Marshal(submission.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal submission data: %w", err)
	}

	metaBytes, err := json.Marshal(submission.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal submission metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query,
		submission.ID,
		submission.FormID,
		dataBytes,
		submission.SubmittedAt,
		submission.Status,
		metaBytes,
	)
	if err != nil {
		return fmt.Errorf("failed to create form submission: %w", err)
	}

	return nil
}

func (s *formSubmissionStore) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	s.logger.Debug("GetByID called", logging.StringField("submission_id", id))
	query := fmt.Sprintf(`SELECT uuid, form_uuid, data, submitted_at, status, metadata 
		FROM form_submissions WHERE uuid = %s`, s.db.GetPlaceholder(1))

	var submission model.FormSubmission
	var dataBytes, metaBytes []byte
	err := s.db.QueryRowxContext(ctx, query, id).Scan(
		&submission.ID,
		&submission.FormID,
		&dataBytes,
		&submission.SubmittedAt,
		&submission.Status,
		&metaBytes,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrFormNotFound
		}
		return nil, fmt.Errorf("failed to get form submission: %w", err)
	}

	if unmarshalErr := json.Unmarshal(dataBytes, &submission.Data); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal submission data: %w", unmarshalErr)
	}

	if unmarshalErr := json.Unmarshal(metaBytes, &submission.Metadata); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal submission metadata: %w", unmarshalErr)
	}

	return &submission, nil
}

func (s *formSubmissionStore) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("GetByFormID called", logging.StringField("form_id", formID))

	query := fmt.Sprintf(`
		SELECT uuid, form_uuid, data, submitted_at, status, metadata
		FROM form_submissions
		WHERE form_uuid = %s
		ORDER BY submitted_at DESC
	`, s.db.GetPlaceholder(1))

	rows, err := s.db.QueryxContext(ctx, query, formID)
	if err != nil {
		return nil, fmt.Errorf("failed to query form submissions: %w", err)
	}
	defer rows.Close()

	var submissions []*model.FormSubmission
	for rows.Next() {
		var submission model.FormSubmission
		var dataBytes, metaBytes []byte
		if scanErr := rows.Scan(
			&submission.ID,
			&submission.FormID,
			&dataBytes,
			&submission.SubmittedAt,
			&submission.Status,
			&metaBytes,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan submission row: %w", scanErr)
		}

		if unmarshalErr := json.Unmarshal(dataBytes, &submission.Data); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal submission data: %w", unmarshalErr)
		}

		if unmarshalErr := json.Unmarshal(metaBytes, &submission.Metadata); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal submission metadata: %w", unmarshalErr)
		}

		submissions = append(submissions, &submission)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating submission rows: %w", rowsErr)
	}

	return submissions, nil
}

func (s *formSubmissionStore) Update(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("Update called", logging.StringField("submission_id", submission.ID))
	query := fmt.Sprintf(`UPDATE form_submissions SET data = %s, status = %s, metadata = %s 
		WHERE uuid = %s`,
		s.db.GetPlaceholder(1),
		s.db.GetPlaceholder(2),
		s.db.GetPlaceholder(3),
		s.db.GetPlaceholder(4),
	)

	dataBytes, err := json.Marshal(submission.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal submission data: %w", err)
	}

	metaBytes, err := json.Marshal(submission.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal submission metadata: %w", err)
	}

	result, err := s.db.ExecContext(ctx, query,
		dataBytes,
		submission.Status,
		metaBytes,
		submission.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update form submission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return model.ErrFormNotFound
	}

	return nil
}

func (s *formSubmissionStore) Delete(ctx context.Context, id string) error {
	s.logger.Debug("Delete called", logging.StringField("submission_id", id))
	query := fmt.Sprintf(`DELETE FROM form_submissions WHERE uuid = %s`, s.db.GetPlaceholder(1))

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete form submission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return model.ErrFormNotFound
	}

	return nil
}
