package form

import (
	"errors"

	"github.com/jonesrussell/goforms/internal/domain/form/model"
)

type service struct {
	store Store
}

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

func (s *service) GetForm(id uint) (*Form, error) {
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

func (s *service) DeleteForm(id uint) error {
	return s.store.Delete(id)
}

func (s *service) UpdateForm(form *Form) error {
	// TODO: Implement form update
	return errors.New("not implemented")
}

// GetFormSubmissions returns all submissions for a form
func (s *service) GetFormSubmissions(formID uint) ([]*model.FormSubmission, error) {
	return s.store.GetFormSubmissions(formID)
}
