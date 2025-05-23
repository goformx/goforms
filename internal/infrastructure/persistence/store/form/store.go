package form

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/jmoiron/sqlx"
)

var (
	ErrFormNotFound = errors.New("form not found")
)

// Store implements form.Store interface
// Implements real DB logic for forms and submissions.
type Store struct {
	db     *sqlx.DB
	logger logging.Logger
}

// NewStore creates a new form store
func NewStore(db *sqlx.DB, logger logging.Logger) form.Store {
	logger.Debug("creating form store",
		logging.BoolField("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

func (s *Store) Create(f *form.Form) error {
	s.logger.Debug("Create called", logging.StringField("form_id", f.ID))
	query := `INSERT INTO forms (uuid, user_id, title, description, schema, active, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	schemaBytes, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	_, err = s.db.Exec(query,
		f.ID,
		f.UserID,
		f.Title,
		f.Description,
		schemaBytes,
		f.Active,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	return nil
}

func (s *Store) GetByID(id string) (*form.Form, error) {
	s.logger.Debug("GetByID called", logging.StringField("form_id", id))
	query := `SELECT uuid, user_id, title, description, schema, active, created_at, updated_at 
		FROM forms WHERE uuid = ?`

	var f form.Form
	var schemaBytes []byte
	err := s.db.QueryRowx(query, id).Scan(
		&f.ID,
		&f.UserID,
		&f.Title,
		&f.Description,
		&schemaBytes,
		&f.Active,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFormNotFound
		}
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	if unmarshalErr := json.Unmarshal(schemaBytes, &f.Schema); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", unmarshalErr)
	}

	return &f, nil
}

func (s *Store) GetByUserID(userID uint) ([]*form.Form, error) {
	s.logger.Debug("GetByUserID called", logging.UintField("user_id", userID))

	query := `
		SELECT uuid, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Queryx(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query forms: %w", err)
	}
	defer rows.Close()

	var forms []*form.Form
	for rows.Next() {
		var f form.Form
		var schemaBytes []byte
		if scanErr := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.Title,
			&f.Description,
			&schemaBytes,
			&f.Active,
			&f.CreatedAt,
			&f.UpdatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan form row: %w", scanErr)
		}

		if unmarshalErr := json.Unmarshal(schemaBytes, &f.Schema); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal form schema: %w", unmarshalErr)
		}
		forms = append(forms, &f)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating form rows: %w", rowsErr)
	}

	return forms, nil
}

func (s *Store) Delete(id string) error {
	s.logger.Debug("Delete called", logging.StringField("form_id", id))
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
		return ErrFormNotFound
	}

	return nil
}

func (s *Store) Update(f *form.Form) error {
	s.logger.Debug("Update called", logging.StringField("form_id", f.ID))
	query := `UPDATE forms SET title = ?, description = ?, schema = ?, active = ?, updated_at = ? 
		WHERE uuid = ?`

	schemaBytes, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	result, err := s.db.Exec(query,
		f.Title,
		f.Description,
		schemaBytes,
		f.Active,
		time.Now(),
		f.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update form: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrFormNotFound
	}

	return nil
}

func (s *Store) GetFormSubmissions(formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("GetFormSubmissions called", logging.StringField("form_id", formID))

	query := `
		SELECT id, form_id, data, submitted_at, status, metadata
		FROM form_submissions
		WHERE form_id = ?
		ORDER BY submitted_at DESC
	`

	rows, err := s.db.Queryx(query, formID)
	if err != nil {
		return nil, fmt.Errorf("failed to query form submissions: %w", err)
	}
	defer rows.Close()

	var submissions []*model.FormSubmission
	for rows.Next() {
		var ssub model.FormSubmission
		var dataBytes, metaBytes []byte
		if scanErr := rows.Scan(
			&ssub.ID,
			&ssub.FormID,
			&dataBytes,
			&ssub.SubmittedAt,
			&ssub.Status,
			&metaBytes,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan submission row: %w", scanErr)
		}

		if dataErr := json.Unmarshal(dataBytes, &ssub.Data); dataErr != nil {
			return nil, fmt.Errorf("failed to unmarshal submission data: %w", dataErr)
		}

		if metaErr := json.Unmarshal(metaBytes, &ssub.Metadata); metaErr != nil {
			return nil, fmt.Errorf("failed to unmarshal submission metadata: %w", metaErr)
		}

		submissions = append(submissions, &ssub)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating submission rows: %w", rowsErr)
	}

	return submissions, nil
}

func (s *Store) CreateSubmission(sub *model.FormSubmission) error {
	s.logger.Debug("CreateSubmission called", logging.StringField("submission_id", sub.ID))
	query := `INSERT INTO form_submissions (id, form_id, data, submitted_at, status, metadata) VALUES (?, ?, ?, ?, ?, ?)`

	dataBytes, err := json.Marshal(sub.Data)
	if err != nil {
		return err
	}

	metaBytes, err := json.Marshal(sub.Metadata)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(query,
		sub.ID,
		sub.FormID,
		dataBytes,
		sub.SubmittedAt,
		sub.Status,
		metaBytes,
	)
	return err
}
