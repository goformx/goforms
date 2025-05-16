package form

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jonesrussell/goforms/internal/domain/form/model"
)

type service struct {
	store Store
}

// NewService creates a new form service instance
func NewService(store Store) Service {
	return &service{
		store: store,
	}
}

func (s *service) CreateForm(userID uint, title, description string, schema JSON) (*Form, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	if schema == nil {
		return nil, errors.New("schema is required")
	}

	form := &Form{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Schema:      schema,
		Active:      true,
	}

	if err := s.store.Create(form); err != nil {
		return nil, err
	}

	return form, nil
}

func (s *service) GetForm(id string) (*Form, error) {
	form, err := s.store.GetByID(id)
	if err != nil {
		return nil, err
	}

	if form == nil {
		return nil, errors.New("form not found")
	}

	return form, nil
}

func (s *service) GetUserForms(userID uint) ([]*Form, error) {
	forms, err := s.store.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return forms, nil
}

func (s *service) DeleteForm(id string) error {
	return s.store.Delete(id)
}

func (s *service) UpdateForm(form *Form) error {
	if form == nil {
		return errors.New("form is required")
	}

	if form.ID == "" {
		return errors.New("form ID is required")
	}

	// Verify the form exists
	existingForm, err := s.store.GetByID(form.ID)
	if err != nil {
		return err
	}

	if existingForm == nil {
		return errors.New("form not found")
	}

	// Update the form in the store
	return s.store.Update(form)
}

// GetFormSubmissions returns all submissions for a form
func (s *service) GetFormSubmissions(formID string) ([]*model.FormSubmission, error) {
	return s.store.GetFormSubmissions(formID)
}
