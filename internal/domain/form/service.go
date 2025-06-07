package form

import (
	"context"

	"github.com/goformx/goforms/internal/domain/common/ctxutil"
	"github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/form/model"
)

type service struct {
	repo      Repository
	publisher event.Publisher
}

// NewService creates a new form service instance
func NewService(repo Repository, publisher event.Publisher) Service {
	return &service{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *service) CreateForm(ctx context.Context, userID uint, title, description string, schema model.JSON) (*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Add user ID to context
	ctx = ctxutil.WithUserID(ctx, userID)

	form := model.NewForm(userID, title, description, schema)

	if err := form.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, form); err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, event.NewFormCreatedEvent(form)); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper error logging
	}

	return form, nil
}

func (s *service) GetForm(ctx context.Context, id string) (*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	form, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if form == nil {
		return nil, model.ErrFormNotFound
	}

	return form, nil
}

func (s *service) GetUserForms(ctx context.Context, userID uint) ([]*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Add user ID to context
	ctx = ctxutil.WithUserID(ctx, userID)

	forms, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return forms, nil
}

func (s *service) DeleteForm(ctx context.Context, id string) error {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.publisher.Publish(ctx, event.NewFormDeletedEvent(id)); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper error logging
	}

	return nil
}

func (s *service) UpdateForm(ctx context.Context, form *model.Form) error {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	if form == nil {
		return model.ErrFormInvalid
	}

	// Add user ID to context
	ctx = ctxutil.WithUserID(ctx, form.UserID)

	if err := form.Validate(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, form); err != nil {
		return err
	}

	if err := s.publisher.Publish(ctx, event.NewFormUpdatedEvent(form)); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper error logging
	}

	return nil
}

// GetFormSubmissions returns all submissions for a form
func (s *service) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	return s.repo.GetFormSubmissions(ctx, formID)
}
