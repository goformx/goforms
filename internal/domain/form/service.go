//go:generate mockgen -typed -source=service.go -destination=../../../test/mocks/form/mock_service.go -package=form -mock_names=Service=MockService

// Package form provides form-related domain services and business logic.
// It includes form creation, validation, submission handling, and related operations.
package form

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/goformx/goforms/internal/domain/common/events"
	formevents "github.com/goformx/goforms/internal/domain/form/events"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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

// UpdateForm updates a form
func (s *formService) UpdateForm(ctx context.Context, form *model.Form) error {
	if validateErr := form.Validate(); validateErr != nil {
		return fmt.Errorf("validate form: %w", validateErr)
	}

	// Update the form
	if updateErr := s.repository.UpdateForm(ctx, form); updateErr != nil {
		return fmt.Errorf("update form in repository: %w", updateErr)
	}

	// Publish form updated event
	event := formevents.NewFormUpdatedEvent(form)
	if publishErr := s.eventBus.Publish(ctx, event); publishErr != nil {
		s.logger.Error("failed to publish form updated event", "error", publishErr)
	}

	return nil
}

// DeleteForm deletes a form
func (s *formService) DeleteForm(ctx context.Context, formID string) error {
	if formID == "" {
		return errors.New("failed to delete form: formID is required")
	}

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
	form, err := s.repository.GetFormByID(ctx, formID)
	if err != nil {
		return nil, fmt.Errorf("get form by ID: %w", err)
	}

	return form, nil
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
	// Validate submission BEFORE any database operations
	if validateErr := submission.Validate(); validateErr != nil {
		return fmt.Errorf("validate form submission: %w", validateErr)
	}

	// Validate the form exists and is active
	form, getErr := s.repository.GetFormByID(ctx, submission.FormID)
	if getErr != nil {
		return fmt.Errorf("get form for submission: %w", getErr)
	}

	if form == nil {
		return errors.New("form not found")
	}

	// Create the submission (validation already passed above)
	if createErr := s.repository.CreateSubmission(ctx, submission); createErr != nil {
		return fmt.Errorf("create form submission: %w", createErr)
	}

	// Publish events for the successfully created submission
	s.publishSubmissionEvents(ctx, submission)

	return nil
}

// publishSubmissionEvents publishes all events related to a form submission
func (s *formService) publishSubmissionEvents(ctx context.Context, submission *model.FormSubmission) {
	// Publish form submitted event
	if err := s.eventBus.Publish(ctx, formevents.NewFormSubmittedEvent(submission)); err != nil {
		s.logger.Error("failed to publish form submitted event", "error", err)
	}

	// Publish validation success event (validation passed before DB write)
	if err := s.eventBus.Publish(ctx, formevents.NewFormValidatedEvent(submission.FormID, true)); err != nil {
		s.logger.Error("failed to publish form validated event", "error", err)
	}

	// Publish form processed event
	if err := s.eventBus.Publish(ctx, formevents.NewFormProcessedEvent(submission.FormID, submission.ID)); err != nil {
		s.logger.Error("failed to publish form processed event", "error", err)
	}
}

// GetFormSubmission retrieves a form submission by ID
func (s *formService) GetFormSubmission(ctx context.Context, submissionID string) (*model.FormSubmission, error) {
	submission, err := s.repository.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("get form submission by ID: %w", err)
	}

	return submission, nil
}

// ListFormSubmissions retrieves a list of form submissions
func (s *formService) ListFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	submissions, err := s.repository.ListSubmissions(ctx, formID)
	if err != nil {
		return nil, fmt.Errorf("list form submissions: %w", err)
	}

	return submissions, nil
}

// UpdateFormState updates the state of a form
func (s *formService) UpdateFormState(ctx context.Context, formID, state string) error {
	form, getErr := s.repository.GetFormByID(ctx, formID)
	if getErr != nil {
		return fmt.Errorf("failed to get form: %w", getErr)
	}

	form.Status = state
	if updateErr := s.repository.UpdateForm(ctx, form); updateErr != nil {
		return fmt.Errorf("failed to update form state: %w", updateErr)
	}

	event := formevents.NewFormStateEvent(formID, state)
	if publishErr := s.eventBus.Publish(ctx, event); publishErr != nil {
		s.logger.Error("failed to publish form state event", "error", publishErr)
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
