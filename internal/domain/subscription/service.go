package subscription

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Service defines the interface for subscription operations
type Service interface {
	CreateSubscription(ctx context.Context, subscription *Subscription) error
	ListSubscriptions(ctx context.Context) ([]Subscription, error)
	GetSubscription(ctx context.Context, id int64) (*Subscription, error)
	GetSubscriptionByEmail(ctx context.Context, email string) (*Subscription, error)
	UpdateSubscriptionStatus(ctx context.Context, id int64, status Status) error
	DeleteSubscription(ctx context.Context, id int64) error
}

// ServiceImpl handles subscription business logic
type ServiceImpl struct {
	store  Store
	logger logging.Logger
}

// NewService creates a new subscription service
func NewService(store Store, logger logging.Logger) Service {
	return &ServiceImpl{
		store:  store,
		logger: logger,
	}
}

// CreateSubscription creates a new subscription
func (s *ServiceImpl) CreateSubscription(ctx context.Context, subscription *Subscription) error {
	if validateErr := s.validateSubscription(ctx, subscription); validateErr != nil {
		s.logger.Error("failed to validate subscription", logging.Error(validateErr))
		return validateErr
	}

	if storeErr := s.store.Create(ctx, subscription); storeErr != nil {
		s.logger.Error("failed to create subscription", logging.Error(storeErr))
		return storeErr
	}

	return nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	// Simple email validation for now
	// In a real application, you might want to use a more robust validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// ListSubscriptions returns all subscriptions
func (s *ServiceImpl) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	subscriptions, err := s.store.List(ctx)
	if err != nil {
		s.logger.Error("failed to list subscriptions",
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return subscriptions, nil
}

// GetSubscription returns a subscription by ID
func (s *ServiceImpl) GetSubscription(ctx context.Context, id int64) (*Subscription, error) {
	subscription, err := s.store.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get subscription",
			logging.Error(err),
		)
		return nil, err
	}

	if subscription == nil {
		s.logger.Error("failed to get subscription",
			logging.Error(ErrSubscriptionNotFound),
		)
		return nil, ErrSubscriptionNotFound
	}

	return subscription, nil
}

// GetSubscriptionByEmail returns a subscription by email
func (s *ServiceImpl) GetSubscriptionByEmail(ctx context.Context, email string) (*Subscription, error) {
	if email == "" {
		return nil, errors.New("invalid input: email is required")
	}

	subscription, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("failed to get subscription by email",
			logging.Error(err),
		)
		return nil, err
	}

	if subscription == nil {
		s.logger.Error("failed to get subscription by email",
			logging.Error(ErrSubscriptionNotFound),
		)
		return nil, ErrSubscriptionNotFound
	}

	return subscription, nil
}

// UpdateSubscriptionStatus updates the status of a subscription
func (s *ServiceImpl) UpdateSubscriptionStatus(ctx context.Context, id int64, status Status) error {
	if validateErr := s.validateStatus(status); validateErr != nil {
		s.logger.Error("failed to validate status", logging.Error(validateErr))
		return validateErr
	}

	if storeErr := s.store.UpdateStatus(ctx, id, status); storeErr != nil {
		s.logger.Error("failed to update subscription status", logging.Error(storeErr))
		return storeErr
	}

	return nil
}

// DeleteSubscription deletes a subscription by ID
func (s *ServiceImpl) DeleteSubscription(ctx context.Context, id int64) error {
	if storeErr := s.store.Delete(ctx, id); storeErr != nil {
		s.logger.Error("failed to delete subscription", logging.Error(storeErr))
		return storeErr
	}

	return nil
}

func (s *ServiceImpl) validateSubscription(ctx context.Context, subscription *Subscription) error {
	// Validate subscription
	if err := subscription.Validate(); err != nil {
		return err
	}

	// Validate email format
	if !isValidEmail(subscription.Email) {
		return ErrInvalidEmail
	}

	// Check if email already exists
	existing, err := s.store.GetByEmail(ctx, subscription.Email)
	if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
		return fmt.Errorf("failed to check existing subscription: %w", err)
	}
	if existing != nil {
		return ErrEmailAlreadyExists
	}

	// Set default values
	subscription.Status = StatusPending
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = time.Now()

	return nil
}

func (s *ServiceImpl) validateStatus(status Status) error {
	if !status.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}
