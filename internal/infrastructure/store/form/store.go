package form

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

type store struct {
	db     *database.Database
	logger logging.Logger
}

func NewStore(db *database.Database, logger logging.Logger) form.Store {
	return &store{
		db:     db,
		logger: logger,
	}
}

func (s *store) Create(f *form.Form) error {
	query := `
		INSERT INTO forms (user_id, title, description, schema, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	schemaJSON, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	err = s.db.QueryRow(
		query,
		f.UserID,
		f.Title,
		f.Description,
		schemaJSON,
		f.Active,
		f.CreatedAt,
		f.UpdatedAt,
	).Scan(&f.ID)

	if err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	return nil
}

func (s *store) GetByID(id uint) (*form.Form, error) {
	query := `
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE id = ?
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
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
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

func (s *store) Delete(id uint) error {
	query := `DELETE FROM forms WHERE id = ?`

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
