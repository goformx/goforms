package form

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var (
	// ErrFormNotFound is returned when a form cannot be found
	ErrFormNotFound = errors.New("form not found")
)

// FormStore implements form.Repository interface
type FormStore struct {
	db     *database.Database
	logger logging.Logger
}

// NewStore creates a new form store
func NewStore(db *database.Database, logger logging.Logger) form.Repository {
	logger.Debug("creating form store",
		logging.BoolField("db_available", db != nil),
	)
	return &FormStore{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form
func (s *FormStore) Create(ctx context.Context, form *model.Form) error {
	// Marshal schema to JSON
	schemaJSON, err := json.Marshal(form.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Create a map for named parameters
	params := map[string]any{
		"uuid":        form.ID,
		"user_id":     form.UserID,
		"title":       form.Title,
		"description": form.Description,
		"schema":      string(schemaJSON),
		"active":      form.Active,
	}

	query := `
		INSERT INTO forms (
			uuid, user_id, title, description, schema, active, created_at, updated_at
		) VALUES (
			:uuid, :user_id, :title, :description, :schema, :active, NOW(), NOW()
		)
	`

	result, err := s.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to insert form: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("failed to insert form: no rows affected")
	}

	return nil
}

// GetByID retrieves a form by its ID
func (s *FormStore) GetByID(ctx context.Context, id string) (*model.Form, error) {
	var form model.Form
	query := `SELECT * FROM forms WHERE uuid = ? AND deleted_at IS NULL`
	err := s.db.GetContext(ctx, &form, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.New(domainerrors.ErrCodeNotFound, "form not found", nil)
		}
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get form")
	}

	// Convert timestamps to UTC
	form.CreatedAt = form.CreatedAt.UTC()
	form.UpdatedAt = form.UpdatedAt.UTC()

	return &form, nil
}

// GetByUserID retrieves all forms created by a specific user
func (s *FormStore) GetByUserID(ctx context.Context, userID uint) ([]*model.Form, error) {
	var forms []*model.Form
	query := `SELECT * FROM forms WHERE user_id = ? AND deleted_at IS NULL ORDER BY created_at DESC`
	err := s.db.SelectContext(ctx, &forms, query, userID)
	if err != nil {
		return nil, domainerrors.Wrap(err, domainerrors.ErrCodeServerError, "failed to get forms by user ID")
	}

	// Convert timestamps to UTC
	for _, form := range forms {
		form.CreatedAt = form.CreatedAt.UTC()
		form.UpdatedAt = form.UpdatedAt.UTC()
	}

	return forms, nil
}

// Update updates an existing form
func (s *FormStore) Update(ctx context.Context, form *model.Form) error {
	// Marshal schema to JSON
	schemaJSON, err := json.Marshal(form.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Create a map for named parameters
	params := map[string]any{
		"uuid":        form.ID,
		"title":       form.Title,
		"description": form.Description,
		"schema":      string(schemaJSON),
		"active":      form.Active,
	}

	query := `
		UPDATE forms SET
			title = :title,
			description = :description,
			schema = :schema,
			active = :active,
			updated_at = NOW()
		WHERE uuid = :uuid
	`

	result, err := s.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update form: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("form not found: %s", form.ID)
	}

	return nil
}

// Delete deletes a form
func (s *FormStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM forms WHERE uuid = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("form not found: %s", id)
	}

	return nil
}

// List retrieves a paginated list of forms
func (s *FormStore) List(ctx context.Context, offset, limit int) ([]*model.Form, error) {
	var forms []*model.Form
	query := `SELECT * FROM forms WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := s.db.SelectContext(ctx, &forms, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list forms: %w", err)
	}

	// Convert timestamps to UTC
	for _, form := range forms {
		form.CreatedAt = form.CreatedAt.UTC()
		form.UpdatedAt = form.UpdatedAt.UTC()
	}

	return forms, nil
}

// Count returns the total number of forms
func (s *FormStore) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM forms WHERE deleted_at IS NULL`
	err := s.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count forms: %w", err)
	}
	return count, nil
}

// Search searches forms by title or description
func (s *FormStore) Search(ctx context.Context, query string, offset, limit int) ([]*model.Form, error) {
	var forms []*model.Form
	sqlQuery := `
		SELECT * FROM forms 
		WHERE deleted_at IS NULL 
		AND (title LIKE ? OR description LIKE ?)
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`
	searchPattern := "%" + query + "%"
	err := s.db.SelectContext(ctx, &forms, sqlQuery, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search forms: %w", err)
	}

	// Convert timestamps to UTC
	for _, form := range forms {
		form.CreatedAt = form.CreatedAt.UTC()
		form.UpdatedAt = form.UpdatedAt.UTC()
	}

	return forms, nil
}

// GetActiveForms returns all active forms
func (s *FormStore) GetActiveForms(ctx context.Context) ([]*model.Form, error) {
	var forms []*model.Form
	query := `SELECT * FROM forms WHERE active = ? AND deleted_at IS NULL ORDER BY created_at DESC`
	err := s.db.SelectContext(ctx, &forms, query, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get active forms: %w", err)
	}

	// Convert timestamps to UTC
	for _, form := range forms {
		form.CreatedAt = form.CreatedAt.UTC()
		form.UpdatedAt = form.UpdatedAt.UTC()
	}

	return forms, nil
}

// GetFormsByStatus returns forms by their active status
func (s *FormStore) GetFormsByStatus(ctx context.Context, active bool) ([]*model.Form, error) {
	var forms []*model.Form
	query := `SELECT * FROM forms WHERE active = ? AND deleted_at IS NULL ORDER BY created_at DESC`
	err := s.db.SelectContext(ctx, &forms, query, active)
	if err != nil {
		return nil, fmt.Errorf("failed to get forms by status: %w", err)
	}

	// Convert timestamps to UTC
	for _, form := range forms {
		form.CreatedAt = form.CreatedAt.UTC()
		form.UpdatedAt = form.UpdatedAt.UTC()
	}

	return forms, nil
}

// GetFormSubmissions gets all submissions for a form
func (s *FormStore) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	query := `SELECT * FROM form_submissions WHERE form_id = ? ORDER BY created_at DESC`
	err := s.db.SelectContext(ctx, &submissions, query, formID)
	if err != nil {
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}

	return submissions, nil
}
