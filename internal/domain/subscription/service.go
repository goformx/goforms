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
		s.logger.Error("failed to check existing subscription",
			logging.String("error", err.Error()),
		)
		return fmt.Errorf("failed to check existing subscription: %w", err)
	}
	if existing != nil {
		return ErrEmailAlreadyExists
	}

	// Set default values
	subscription.Status = StatusPending
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = time.Now()

	// Create subscription
	if err := s.store.Create(ctx, subscription); err != nil {
		s.logger.Error("failed to create subscription",
			logging.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create subscription: %w", err)
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
			logging.String("error", err.Error()),
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
			logging.String("error", err.Error()),
		)
		return nil, err
	}

	if subscription == nil {
		s.logger.Error("failed to get subscription",
			logging.String("error", ErrSubscriptionNotFound.Error()),
		)
		return nil, ErrSubscriptionNotFound
	}

	return subscription, nil
}

// GetSubscriptionByEmail returns a subscription by email
func (s *ServiceImpl) GetSubscriptionByEmail(ctx context.Context, email string) (*Subscription, error) {
	if email == "" {
		return nil, fmt.Errorf("invalid input: email is required")
	}

	subscription, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("failed to get subscription by email",
			logging.String("error", err.Error()),
		)
		return nil, err
	}

	if subscription == nil {
		s.logger.Error("failed to get subscription by email",
			logging.String("error", ErrSubscriptionNotFound.Error()),
		)
		return nil, ErrSubscriptionNotFound
	}

	return subscription, nil
}

// UpdateSubscriptionStatus updates the status of a subscription
func (s *ServiceImpl) UpdateSubscriptionStatus(ctx context.Context, id int64, status Status) error {
	if !status.IsValid() {
		return ErrInvalidStatus
	}

	// Check if subscription exists
	subscription, err := s.store.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get subscription",
			logging.String("error", err.Error()),
		)
		return err
	}

	if subscription == nil {
		s.logger.Error("failed to get subscription",
			logging.String("error", ErrSubscriptionNotFound.Error()),
		)
		return ErrSubscriptionNotFound
	}

	// Update status
	if err := s.store.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("failed to update subscription status",
			logging.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// DeleteSubscription removes a subscription
func (s *ServiceImpl) DeleteSubscription(ctx context.Context, id int64) error {
	// Check if subscription exists
	subscription, err := s.store.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get subscription",
			logging.String("error", err.Error()),
		)
		return err
	}

	if subscription == nil {
		s.logger.Error("failed to get subscription",
			logging.String("error", ErrSubscriptionNotFound.Error()),
		)
		return ErrSubscriptionNotFound
	}

	// Delete subscription
	if err := s.store.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete subscription",
			logging.String("error", err.Error()),
		)
		return err
	}

	return nil
}
