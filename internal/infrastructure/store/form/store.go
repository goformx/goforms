package form

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

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

func (s *store) Create(f *form.Form) error {
	query := `
		INSERT INTO forms (uuid, user_id, title, description, schema, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	schemaJSON, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	_, err = s.db.Exec(
		query,
		f.ID,
		f.UserID,
		f.Title,
		f.Description,
		schemaJSON,
		f.Active,
		f.CreatedAt,
		f.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create form: %w", err)
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

	err := s.db.QueryRow(query, id).Scan(
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
			return nil, errors.New("form not found")
		}
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	if unmarshalErr := json.Unmarshal(schemaJSON, &f.Schema); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", unmarshalErr)
	}

	return f, nil
}

func (s *store) GetByUserID(userID uint) ([]*form.Form, error) {
	query := `
		SELECT uuid, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE user_id = ?
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query forms: %w", err)
	}
	defer rows.Close()

	var forms []*form.Form
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
			log.Printf("Error scanning form row: %v", scanErr)
			continue
		}

		if unmarshalErr := json.Unmarshal(schemaJSON, &f.Schema); unmarshalErr != nil {
			log.Printf("Error unmarshaling schema: %v", unmarshalErr)
			continue
		}

		forms = append(forms, f)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating forms: %w", rowsErr)
	}

	return forms, nil
}

func (s *store) Delete(id string) error {
	query := `DELETE FROM forms WHERE uuid = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("form not found")
	}

	return nil
}

func (s *store) GetFormSubmissions(formID string) ([]*model.FormSubmission, error) {
	query := `
		SELECT id, form_uuid, data, submitted_at, status, metadata
		FROM form_submissions
		WHERE form_uuid = ?
	`

	rows, queryErr := s.db.Query(query, formID)
	if queryErr != nil {
		return nil, fmt.Errorf("failed to query form submissions: %w", queryErr)
	}
	defer rows.Close()

	var submissions []*model.FormSubmission
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
			log.Printf("Error scanning submission row: %v", scanErr)
			continue
		}

		if unmarshalErr := json.Unmarshal(dataJSON, &submission.Data); unmarshalErr != nil {
			log.Printf("Error unmarshaling submission data: %v", unmarshalErr)
			continue
		}

		if metadataErr := json.Unmarshal(metadataJSON, &submission.Metadata); metadataErr != nil {
			log.Printf("Error unmarshaling submission metadata: %v", metadataErr)
			continue
		}

		submissions = append(submissions, &submission)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating submissions: %w", rowsErr)
	}

	return submissions, nil
}

func (s *store) Update(f *form.Form) error {
	// Marshal the schema to JSON
	schemaJSON, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Update the form
	query := `
		UPDATE forms 
		SET title = ?, description = ?, schema = ?, active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE uuid = ?
	`

	result, err := s.db.Exec(query, f.Title, f.Description, schemaJSON, f.Active, f.ID)
	if err != nil {
		return fmt.Errorf("failed to update form: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("form not found")
	}

	return nil
}
