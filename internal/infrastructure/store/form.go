package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// ErrFormNotFound is returned when a form is not found
var ErrFormNotFound = errors.New("form not found")

type FormStore struct {
	db     *database.Database
	logger logging.Logger
}

func NewFormStore(db *database.Database, logger logging.Logger) form.Store {
	return &FormStore{
		db:     db,
		logger: logger,
	}
}

func (s *FormStore) Create(f *form.Form) error {
	query := `
		INSERT INTO forms (user_id, title, description, schema, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := s.db.QueryRow(
		query,
		f.UserID,
		f.Title,
		f.Description,
		f.Schema,
		f.Active,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	// Check for integer overflow
	if id < 0 || id > math.MaxUint32 {
		return fmt.Errorf("form ID out of range: %d", id)
	}

	f.ID = uint(id)
	return nil
}

func (s *FormStore) GetByID(id uint) (*form.Form, error) {
	query := `
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE id = ?
	`

	var f form.Form
	var schemaJSON []byte
	var createdAt, updatedAt string

	err := s.db.QueryRow(query, id).Scan(
		&f.ID,
		&f.UserID,
		&f.Title,
		&f.Description,
		&schemaJSON,
		&f.Active,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFormNotFound
		}
		return nil, err
	}

	unmarshalErr := json.Unmarshal(schemaJSON, &f.Schema)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	f.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	f.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &f, nil
}

func (s *FormStore) GetByUserID(userID uint) ([]*form.Form, error) {
	query := `
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, queryErr := s.db.Query(query, userID)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	var forms []*form.Form
	for rows.Next() {
		var f form.Form
		var schemaJSON []byte
		var createdAt, updatedAt string

		scanErr := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.Title,
			&f.Description,
			&schemaJSON,
			&f.Active,
			&createdAt,
			&updatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}

		unmarshalErr := json.Unmarshal(schemaJSON, &f.Schema)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}

		f.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		f.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		forms = append(forms, &f)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", rowsErr)
	}

	return forms, nil
}

func (s *FormStore) Update(f *form.Form) error {
	query := `
		UPDATE forms
		SET title = ?, description = ?, schema = ?, active = ?
		WHERE id = ? AND user_id = ?
	`

	schemaJSON, err := json.Marshal(f.Schema)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(query, f.Title, f.Description, schemaJSON, f.Active, f.ID, f.UserID)
	return err
}

func (s *FormStore) Delete(id uint) error {
	query := `DELETE FROM forms WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}
