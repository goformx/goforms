package form

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/common/events"
	formevents "github.com/goformx/goforms/internal/domain/form/events"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/google/uuid"
)

const (
	// DefaultTimeout is the default timeout for form operations
	DefaultTimeout = 30 * time.Second
)

// Service defines the interface for form-related business logic
type Service interface {
	CreateForm(ctx context.Context, form *model.Form) error
	UpdateForm(ctx context.Context, form *model.Form) error
	DeleteForm(ctx context.Context, formID string) error
	GetForm(ctx context.Context, formID string) (*model.Form, error)
	ListForms(ctx context.Context, userID string) ([]*model.Form, error)
	SubmitForm(ctx context.Context, submission *model.FormSubmission) error
	GetFormSubmission(ctx context.Context, submissionID string) (*model.FormSubmission, error)
	ListFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error)
	UpdateFormState(ctx context.Context, formID, state string) error
	TrackFormAnalytics(ctx context.Context, formID, eventType string) error
}

// formService handles form-related business logic
type formService struct {
	repository Repository
	eventBus   events.EventBus
	logger     logging.Logger
}

// NewService creates a new form service
func NewService(repository Repository, eventBus events.EventBus, logger logging.Logger) Service {
	return &formService{
		repository: repository,
		eventBus:   eventBus,
		logger:     logger,
	}
}

// CreateForm creates a new form
func (s *formService) CreateForm(ctx context.Context, form *model.Form) error {
	if err := form.Validate(); err != nil {
		return fmt.Errorf("form validation failed: %w", err)
	}

	// Set form ID if not already set
	if form.ID == "" {
		form.ID = uuid.New().String()
	}

	if err := s.repository.CreateForm(ctx, form); err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	if err := s.eventBus.Publish(ctx, formevents.NewFormCreatedEvent(form)); err != nil {
		s.logger.Error("failed to publish form created event", "error", err)
	}

	return nil
}

// UpdateForm updates an existing form
func (s *formService) UpdateForm(ctx context.Context, form *model.Form) error {
	if err := form.Validate(); err != nil {
		return fmt.Errorf("invalid form: %w", err)
	}

	updateErr := s.repository.UpdateForm(ctx, form)
	if updateErr != nil {
		return fmt.Errorf("failed to update form: %w", updateErr)
	}

	event := formevents.NewFormUpdatedEvent(form)
	publishErr := s.eventBus.Publish(ctx, event)
	if publishErr != nil {
		return fmt.Errorf("failed to publish update event: %w", publishErr)
	}

	return nil
}

// DeleteForm deletes a form
func (s *formService) DeleteForm(ctx context.Context, formID string) error {
	if err := s.repository.DeleteForm(ctx, formID); err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
	}

	if err := s.eventBus.Publish(ctx, formevents.NewFormDeletedEvent(formID)); err != nil {
		s.logger.Error("failed to publish form deleted event", "error", err)
	}

	return nil
}

// GetForm retrieves a form by ID
func (s *formService) GetForm(ctx context.Context, formID string) (*model.Form, error) {
	return s.repository.GetFormByID(ctx, formID)
}

// ListForms retrieves a list of forms
func (s *formService) ListForms(ctx context.Context, userID string) ([]*model.Form, error) {
	forms, err := s.repository.ListForms(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list forms: %w", err)
	}
	return forms, nil
}

// SubmitForm submits a form
func (s *formService) SubmitForm(ctx context.Context, submission *model.FormSubmission) error {
	if err := submission.Validate(); err != nil {
		return err
	}

	// Validate the form exists
	form, err := s.repository.GetFormByID(ctx, submission.FormID)
	if err != nil {
		return err
	}
	if form == nil {
		return errors.New("form not found")
	}

	// Create the submission
	createErr := s.repository.CreateSubmission(ctx, submission)
	if createErr != nil {
		return createErr
	}

	// Publish form submitted event
	event := formevents.NewFormSubmittedEvent(submission)
	publishErr := s.eventBus.Publish(ctx, event)
	if publishErr != nil {
		s.logger.Error("failed to publish form submitted event", "error", publishErr)
	}

	// Validate the submission
	validationErr := submission.Validate()
	isValid := validationErr == nil
	validationEvent := formevents.NewFormValidatedEvent(submission.FormID, isValid)
	validationPublishErr := s.eventBus.Publish(ctx, validationEvent)
	if validationPublishErr != nil {
		s.logger.Error("failed to publish form validated event", "error", validationPublishErr)
	}

	if !isValid {
		return validationErr
	}

	processingEvent := formevents.NewFormProcessedEvent(submission.FormID, submission.ID)
	processingPublishErr := s.eventBus.Publish(ctx, processingEvent)
	if processingPublishErr != nil {
		s.logger.Error("failed to publish form processed event", "error", processingPublishErr)
	}

	return nil
}

// GetFormSubmission retrieves a form submission by ID
func (s *formService) GetFormSubmission(ctx context.Context, submissionID string) (*model.FormSubmission, error) {
	return s.repository.GetSubmissionByID(ctx, submissionID)
}

// ListFormSubmissions retrieves a list of form submissions
func (s *formService) ListFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	return s.repository.ListSubmissions(ctx, formID)
}

// UpdateFormState updates the state of a form
func (s *formService) UpdateFormState(ctx context.Context, formID, state string) error {
	form, err := s.repository.GetFormByID(ctx, formID)
	if err != nil {
		return fmt.Errorf("failed to get form: %w", err)
	}

	form.Status = state
	if err := s.repository.UpdateForm(ctx, form); err != nil {
		return fmt.Errorf("failed to update form state: %w", err)
	}

	event := formevents.NewFormStateEvent(formID, state)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("failed to publish form state event", "error", err)
	}

	return nil
}

// TrackFormAnalytics tracks form analytics
func (s *formService) TrackFormAnalytics(ctx context.Context, formID, eventType string) error {
	event := formevents.NewAnalyticsEvent(formID, eventType)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish analytics event: %w", err)
	}
	return nil
}
