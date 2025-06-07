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

var (
	ErrFormNotFound = errors.New("form not found")
)

// formStore implements form.Repository interface
type formStore struct {
	db     *database.Database
	logger logging.Logger
}

// NewStore creates a new form store
func NewStore(db *database.Database, logger logging.Logger) form.Repository {
	logger.Debug("creating form store",
		logging.BoolField("db_available", db != nil),
	)
	return &formStore{
		db:     db,
		logger: logger,
	}
}

func (s *formStore) Create(ctx context.Context, f *model.Form) error {
	s.logger.Debug("Create called", logging.StringField("form_id", f.ID))
	query := fmt.Sprintf(`INSERT INTO forms (uuid, user_id, title, description, schema, active, created_at, updated_at) 
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s)`,
		s.db.GetPlaceholder(1),
		s.db.GetPlaceholder(2),
		s.db.GetPlaceholder(3),
		s.db.GetPlaceholder(4),
		s.db.GetPlaceholder(5),
		s.db.GetPlaceholder(6),
		s.db.GetPlaceholder(7),
		s.db.GetPlaceholder(8),
	)

	schemaBytes, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query,
		f.ID,
		f.UserID,
		f.Title,
		f.Description,
		schemaBytes,
		f.Active,
		f.CreatedAt,
		f.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	return nil
}

func (s *formStore) GetByID(ctx context.Context, id string) (*model.Form, error) {
	s.logger.Debug("GetByID called", logging.StringField("form_id", id))
	query := fmt.Sprintf(`SELECT uuid, user_id, title, description, schema, active, created_at, updated_at 
		FROM forms WHERE uuid = %s`, s.db.GetPlaceholder(1))

	var f model.Form
	var schemaBytes []byte
	err := s.db.QueryRowxContext(ctx, query, id).Scan(
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
			return nil, model.ErrFormNotFound
		}
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	if unmarshalErr := json.Unmarshal(schemaBytes, &f.Schema); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", unmarshalErr)
	}

	return &f, nil
}

func (s *formStore) GetByUserID(ctx context.Context, userID uint) ([]*model.Form, error) {
	s.logger.Debug("GetByUserID called", logging.UintField("user_id", userID))

	query := fmt.Sprintf(`
		SELECT uuid, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE user_id = %s
		ORDER BY created_at DESC
	`, s.db.GetPlaceholder(1))

	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query forms: %w", err)
	}
	defer rows.Close()

	var forms []*model.Form
	for rows.Next() {
		var f model.Form
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

func (s *formStore) Delete(ctx context.Context, id string) error {
	s.logger.Debug("Delete called", logging.StringField("form_id", id))
	query := fmt.Sprintf(`DELETE FROM forms WHERE uuid = %s`, s.db.GetPlaceholder(1))

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
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

func (s *formStore) Update(ctx context.Context, f *model.Form) error {
	s.logger.Debug("Update called", logging.StringField("form_id", f.ID))
	query := fmt.Sprintf(`UPDATE forms SET title = %s, description = %s, schema = %s, active = %s, updated_at = %s 
		WHERE uuid = %s`,
		s.db.GetPlaceholder(1),
		s.db.GetPlaceholder(2),
		s.db.GetPlaceholder(3),
		s.db.GetPlaceholder(4),
		s.db.GetPlaceholder(5),
		s.db.GetPlaceholder(6),
	)

	schemaBytes, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	result, err := s.db.ExecContext(ctx, query,
		f.Title,
		f.Description,
		schemaBytes,
		f.Active,
		f.UpdatedAt,
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
		return model.ErrFormNotFound
	}

	return nil
}

func (s *formStore) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("GetFormSubmissions called", logging.StringField("form_id", formID))

	query := fmt.Sprintf(`
		SELECT id, form_uuid, data, submitted_at, status, metadata
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

func (s *formStore) CreateSubmission(sub *model.FormSubmission) error {
	s.logger.Debug("CreateSubmission called", logging.StringField("submission_id", sub.ID))
	query := `INSERT INTO form_submissions (id, form_uuid, data, submitted_at, status, metadata) VALUES (?, ?, ?, ?, ?, ?)`

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
