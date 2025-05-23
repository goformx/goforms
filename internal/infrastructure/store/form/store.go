package form

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/form/model"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

type store struct {
	db     *database.Database
	logger logging.Logger
}

// NewStore creates a new form store instance
func NewStore(db *database.Database, logger logging.Logger) form.Store {
	return &store{
		db:     db,
		logger: logger,
	}
}

// Helper functions for database operations
func (s *store) execQueryWithArgs(query string, operation string, args []interface{}, fields ...any) (sql.Result, error) {
	result, err := s.db.Exec(query, args...)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to %s", operation),
			append(fields, "error", err)...,
		)
		return nil, fmt.Errorf("database operation failed: %w", err)
	}
	return result, nil
}

func (s *store) queryRow(query string, operation string, args []interface{}, fields ...any) *sql.Row {
	row := s.db.QueryRow(query, args...)
	if row.Err() != nil {
		s.logger.Error(fmt.Sprintf("Failed to %s", operation),
			append(fields, "error", row.Err())...,
		)
	}
	return row
}

func (s *store) query(query string, operation string, args []interface{}, fields ...any) (*sql.Rows, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to %s", operation),
			append(fields, "error", err)...,
		)
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	return rows, nil
}

// Helper function for JSON validation and marshaling
func (s *store) marshalJSON(data interface{}, context string, id string) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Failed to marshal JSON",
			"error", err,
			"context", context,
			"id", id,
		)
		return nil, fmt.Errorf("failed to marshal %s: %w", context, err)
	}

	if !json.Valid(jsonData) {
		s.logger.Error("Invalid JSON produced",
			"context", context,
			"id", id,
		)
		return nil, fmt.Errorf("invalid JSON produced for %s", context)
	}

	return jsonData, nil
}

// Helper function for JSON validation and unmarshaling
func (s *store) unmarshalJSON(data []byte, target interface{}, context string, id string) error {
	if !json.Valid(data) {
		s.logger.Error("Invalid JSON received",
			"context", context,
			"id", id,
		)
		return fmt.Errorf("invalid JSON received for %s", context)
	}

	if err := json.Unmarshal(data, target); err != nil {
		s.logger.Error("Failed to unmarshal JSON",
			"error", err,
			"context", context,
			"id", id,
		)
		return fmt.Errorf("failed to unmarshal %s: %w", context, err)
	}

	return nil
}

func (s *store) Create(f *form.Form) error {
	query := `
		INSERT INTO forms (uuid, user_id, title, description, schema, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	schemaJSON, err := s.marshalJSON(f.Schema, "form schema", f.ID)
	if err != nil {
		return err
	}

	args := []interface{}{
		f.ID,
		f.UserID,
		f.Title,
		f.Description,
		schemaJSON,
		f.Active,
		f.CreatedAt,
		f.UpdatedAt,
	}

	_, err = s.execQueryWithArgs(query, "create form", args, "form_id", f.ID)
	if err != nil {
		return fmt.Errorf("failed to create form with ID %s: %w", f.ID, err)
	}

	return nil
}

func (s *store) GetByID(id string) (*form.Form, error) {
	query := `
		SELECT uuid, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE uuid = ?
	`

	var schemaJSON []byte
	f := &form.Form{}

	err := s.queryRow(query, "get form", []interface{}{id}, "form_id", id).Scan(
		&f.ID,
		&f.UserID,
		&f.Title,
		&f.Description,
		&schemaJSON,
		&f.Active,
		&f.CreatedAt,
		&f.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Info("Form not found", "form_id", id)
			return nil, fmt.Errorf("form with ID %s not found", id)
		}
		s.logger.Error("Failed to get form", "error", err, "form_id", id)
		return nil, fmt.Errorf("failed to retrieve form with ID %s: %w", id, err)
	}

	if err := s.unmarshalJSON(schemaJSON, &f.Schema, "form schema", id); err != nil {
		return nil, err
	}

	return f, nil
}

func (s *store) GetByUserID(userID uint) ([]*form.Form, error) {
	query := `
		SELECT uuid, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE user_id = ?
	`

	rows, err := s.query(query, "query forms", []interface{}{userID}, "user_id", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query forms for user %d: %w", userID, err)
	}
	defer rows.Close()

	var forms []*form.Form
	var errList []error

	for rows.Next() {
		var schemaJSON []byte
		f := &form.Form{}

		scanErr := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.Title,
			&f.Description,
			&schemaJSON,
			&f.Active,
			&f.CreatedAt,
			&f.UpdatedAt,
		)
		if scanErr != nil {
			errList = append(errList, fmt.Errorf("failed to scan form data for user %d: %w", userID, scanErr))
			continue
		}

		if err := s.unmarshalJSON(schemaJSON, &f.Schema, "form schema", f.ID); err != nil {
			errList = append(errList, err)
			continue
		}

		forms = append(forms, f)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		errList = append(errList, fmt.Errorf("error while processing forms for user %d: %w", userID, rowsErr))
	}

	if len(errList) > 0 {
		s.logger.Error("Multiple errors occurred while processing forms",
			"user_id", userID,
			"error_count", len(errList),
			"errors", errList,
		)
		return forms, fmt.Errorf("multiple errors occurred while processing forms for user %d: %v", userID, errList)
	}

	return forms, nil
}

func (s *store) Delete(id string) error {
	query := `DELETE FROM forms WHERE uuid = ?`

	result, err := s.execQueryWithArgs(query, "delete form", []interface{}{id}, "form_id", id)
	if err != nil {
		return fmt.Errorf("failed to delete form with ID %s: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", "error", err, "form_id", id)
		return fmt.Errorf("failed to verify form deletion for ID %s: %w", id, err)
	}

	if rowsAffected == 0 {
		s.logger.Info("Form not found for deletion", "form_id", id)
		return fmt.Errorf("form with ID %s not found for deletion", id)
	}

	return nil
}

func (s *store) GetFormSubmissions(formID string) ([]*model.FormSubmission, error) {
	query := `
		SELECT id, form_uuid, data, submitted_at, status, metadata
		FROM form_submissions
		WHERE form_uuid = ?
	`

	rows, err := s.query(query, "query form submissions", []interface{}{formID}, "form_id", formID)
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions for form %s: %w", formID, err)
	}
	defer rows.Close()

	var submissions []*model.FormSubmission
	var errList []error

	for rows.Next() {
		var submission model.FormSubmission
		var dataJSON, metadataJSON []byte

		scanErr := rows.Scan(
			&submission.ID,
			&submission.FormID,
			&dataJSON,
			&submission.SubmittedAt,
			&submission.Status,
			&metadataJSON,
		)
		if scanErr != nil {
			errList = append(errList, fmt.Errorf("failed to scan submission data for form %s: %w", formID, scanErr))
			continue
		}

		if err := s.unmarshalJSON(dataJSON, &submission.Data, "submission data", submission.ID); err != nil {
			errList = append(errList, err)
			continue
		}

		if err := s.unmarshalJSON(metadataJSON, &submission.Metadata, "submission metadata", submission.ID); err != nil {
			errList = append(errList, err)
			continue
		}

		submissions = append(submissions, &submission)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		errList = append(errList, fmt.Errorf("error while processing submissions for form %s: %w", formID, rowsErr))
	}

	if len(errList) > 0 {
		s.logger.Error("Multiple errors occurred while processing submissions",
			"form_id", formID,
			"error_count", len(errList),
			"errors", errList,
		)
		return submissions, fmt.Errorf("multiple errors occurred while processing submissions for form %s: %v", formID, errList)
	}

	return submissions, nil
}

func (s *store) Update(f *form.Form) error {
	// Marshal the schema to JSON
	schemaJSON, err := s.marshalJSON(f.Schema, "form schema", f.ID)
	if err != nil {
		return err
	}

	// Update the form
	query := `
		UPDATE forms 
		SET title = ?, description = ?, schema = ?, active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE uuid = ?
	`

	args := []interface{}{
		f.Title,
		f.Description,
		schemaJSON,
		f.Active,
		f.ID,
	}

	result, err := s.execQueryWithArgs(query, "update form", args, "form_id", f.ID)
	if err != nil {
		return fmt.Errorf("failed to update form with ID %s: %w", f.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", "error", err, "form_id", f.ID)
		return fmt.Errorf("failed to verify form update for ID %s: %w", f.ID, err)
	}

	if rowsAffected == 0 {
		s.logger.Info("Form not found for update", "form_id", f.ID)
		return fmt.Errorf("form with ID %s not found for update", f.ID)
	}

	return nil
}
