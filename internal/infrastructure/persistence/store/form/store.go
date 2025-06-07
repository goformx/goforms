package form

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var (
	ErrFormNotFound = errors.New("form not found")
	ErrFormInvalid  = errors.New("invalid form data")
)

// formModel represents the database model for forms
type formModel struct {
	ID          string    `db:"uuid"`
	UserID      uint      `db:"user_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Schema      string    `db:"schema"`
	Active      bool      `db:"active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// submissionModel represents the database model for form submissions
type submissionModel struct {
	ID        string    `db:"id"`
	FormID    string    `db:"form_uuid"`
	Data      string    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// FormStore implements form.Repository
type FormStore struct {
	db     *database.Database
	logger logging.Logger
}

// NewFormStore creates a new form store
func NewFormStore(db *database.Database, logger logging.Logger) *FormStore {
	return &FormStore{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form
func (s *FormStore) Create(ctx context.Context, form *model.Form) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "create_form"),
				logging.Error(err),
			)
		}
	}()

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

	result, err := tx.NamedExecContext(ctx, query, params)
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a form by ID
func (s *FormStore) GetByID(ctx context.Context, id string) (*model.Form, error) {
	var form formModel
	query := fmt.Sprintf(`SELECT * FROM forms WHERE uuid = %s`, s.db.GetPlaceholder(1))
	err := s.db.GetContext(ctx, &form, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("form not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get form: %w", err)
	}
	return s.mapToForm(&form), nil
}

// GetByUserID retrieves forms by user ID
func (s *FormStore) GetByUserID(ctx context.Context, userID uint) ([]*model.Form, error) {
	var forms []formModel
	query := fmt.Sprintf(`SELECT * FROM forms WHERE user_id = %s`, s.db.GetPlaceholder(1))
	err := s.db.SelectContext(ctx, &forms, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get forms: %w", err)
	}
	return s.mapToForms(forms), nil
}

// Update updates an existing form
func (s *FormStore) Update(ctx context.Context, form *model.Form) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "update_form"),
				logging.Error(err),
			)
		}
	}()

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

	result, err := tx.NamedExecContext(ctx, query, params)
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete deletes a form
func (s *FormStore) Delete(ctx context.Context, id string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			s.logger.Error("failed to rollback transaction",
				logging.String("operation", "delete_form"),
				logging.Error(err),
			)
		}
	}()

	query := fmt.Sprintf(`DELETE FROM forms WHERE uuid = %s`, s.db.GetPlaceholder(1))

	result, err := tx.ExecContext(ctx, query, id)
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// List returns a paginated list of forms
func (s *FormStore) List(ctx context.Context, offset, limit int) ([]*model.Form, error) {
	var forms []formModel
	query := `SELECT * FROM forms ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := s.db.SelectContext(ctx, &forms, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list forms: %w", err)
	}
	return s.mapToForms(forms), nil
}

// Count returns the total number of forms
func (s *FormStore) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM forms")
	if err != nil {
		return 0, fmt.Errorf("failed to count forms: %w", err)
	}
	return count, nil
}

// Search searches forms by title or description
func (s *FormStore) Search(ctx context.Context, query string, offset, limit int) ([]*model.Form, error) {
	var forms []formModel
	sqlQuery := `
		SELECT * FROM forms 
		WHERE title LIKE ? OR description LIKE ?
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`
	searchPattern := "%" + query + "%"
	err := s.db.SelectContext(ctx, &forms, sqlQuery, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search forms: %w", err)
	}
	return s.mapToForms(forms), nil
}

// GetActiveForms returns all active forms
func (s *FormStore) GetActiveForms(ctx context.Context) ([]*model.Form, error) {
	var forms []formModel
	query := `SELECT * FROM forms WHERE active = true ORDER BY created_at DESC`
	err := s.db.SelectContext(ctx, &forms, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active forms: %w", err)
	}
	return s.mapToForms(forms), nil
}

// GetFormsByStatus returns forms by their active status
func (s *FormStore) GetFormsByStatus(ctx context.Context, active bool) ([]*model.Form, error) {
	var forms []formModel
	query := fmt.Sprintf(`SELECT * FROM forms WHERE active = %s`, s.db.GetPlaceholder(1))
	err := s.db.SelectContext(ctx, &forms, query, active)
	if err != nil {
		return nil, fmt.Errorf("failed to get forms: %w", err)
	}
	return s.mapToForms(forms), nil
}

// GetFormSubmissions gets all submissions for a form
func (s *FormStore) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	var submissions []submissionModel
	query := fmt.Sprintf(`SELECT * FROM form_submissions WHERE form_uuid = %s ORDER BY created_at DESC`, s.db.GetPlaceholder(1))
	err := s.db.SelectContext(ctx, &submissions, query, formID)
	if err != nil {
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}
	return s.mapToSubmissions(submissions), nil
}

// mapToForm converts a formModel to a model.Form
func (s *FormStore) mapToForm(f *formModel) *model.Form {
	var schema model.JSON
	if err := json.Unmarshal([]byte(f.Schema), &schema); err != nil {
		s.logger.Error("failed to unmarshal form schema",
			logging.String("operation", "map_form"),
			logging.Error(err),
		)
		schema = model.JSON{}
	}

	return &model.Form{
		ID:          f.ID,
		UserID:      f.UserID,
		Title:       f.Title,
		Description: f.Description,
		Schema:      schema,
		Active:      f.Active,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

// mapToForms converts a slice of formModel to a slice of model.Form
func (s *FormStore) mapToForms(forms []formModel) []*model.Form {
	result := make([]*model.Form, len(forms))
	for i, f := range forms {
		result[i] = s.mapToForm(&f)
	}
	return result
}

// mapToSubmission converts a submissionModel to a model.FormSubmission
func (s *FormStore) mapToSubmission(sm *submissionModel) *model.FormSubmission {
	var data model.JSON
	if err := json.Unmarshal([]byte(sm.Data), &data); err != nil {
		s.logger.Error("failed to unmarshal submission data",
			logging.String("operation", "map_submission"),
			logging.Error(err),
		)
		data = model.JSON{}
	}

	return &model.FormSubmission{
		ID:          sm.ID,
		FormID:      sm.FormID,
		Data:        data,
		SubmittedAt: sm.CreatedAt,
		Status:      model.SubmissionStatusPending,
		Metadata:    model.JSON{},
	}
}

// mapToSubmissions converts a slice of submissionModel to a slice of model.FormSubmission
func (s *FormStore) mapToSubmissions(submissions []submissionModel) []*model.FormSubmission {
	result := make([]*model.FormSubmission, len(submissions))
	for i, sm := range submissions {
		result[i] = s.mapToSubmission(&sm)
	}
	return result
}
