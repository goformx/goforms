package contact

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Ensure Service implements ContactService interface
var _ Service = (*ServiceImpl)(nil)

// ServiceImpl handles contact form business logic
type ServiceImpl struct {
	store  Store
	logger logging.Logger
}

// NewService creates a new contact service
func NewService(store Store, log logging.Logger) *ServiceImpl {
	return &ServiceImpl{
		store:  store,
		logger: log,
	}
}

// Submit handles a new contact form submission
func (s *ServiceImpl) Submit(ctx context.Context, sub *Submission) error {
	if sub.Name == "" {
		return ErrNameRequired
	}
	if sub.Email == "" {
		return ErrEmailRequired
	}
	if sub.Message == "" {
		return ErrMessageRequired
	}

	sub.Status = StatusPending
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = sub.CreatedAt

	if err := s.store.Create(ctx, sub); err != nil {
		s.logger.Error("failed to create submission",
			logging.Error(err),
			logging.String("email", sub.Email),
		)
		return fmt.Errorf("failed to create submission: %w", err)
	}

	s.logger.Info("submission created",
		logging.String("email", sub.Email),
		logging.String("status", string(sub.Status)),
	)

	return nil
}

// ListSubmissions returns all contact form submissions
func (s *ServiceImpl) ListSubmissions(ctx context.Context) ([]Submission, error) {
	submissions, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}
	return submissions, nil
}

// GetSubmission returns a specific contact form submission
func (s *ServiceImpl) GetSubmission(ctx context.Context, id int64) (*Submission, error) {
	submission, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}
	return submission, nil
}

// UpdateSubmissionStatus updates the status of a contact form submission
func (s *ServiceImpl) UpdateSubmissionStatus(ctx context.Context, id int64, status Status) error {
	if status != StatusPending && status != StatusApproved && status != StatusRejected {
		return ErrInvalidStatus
	}

	if err := s.store.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("failed to update submission status",
			logging.Error(err),
			logging.String("id", strconv.FormatInt(id, 10)),
			logging.String("status", string(status)),
		)
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	s.logger.Info("submission status updated",
		logging.String("id", strconv.FormatInt(id, 10)),
		logging.String("status", string(status)),
	)

	return nil
}
