package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/goforms/internal/logger"
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
	log   logger.Logger
	store Store
}

// NewService creates a new subscription service
func NewService(log logger.Logger, store Store) Service {
	return &ServiceImpl{
		log:   log,
		store: store,
	}
}

func (s *ServiceImpl) wrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// CreateSubscription creates a new subscription
func (s *ServiceImpl) CreateSubscription(ctx context.Context, subscription *Subscription) error {
	// Check if email already exists
	existing, err := s.store.GetByEmail(ctx, subscription.Email)
	if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
		s.log.Error("failed to check existing subscription", logger.Error(err))
		return s.wrapError(err, "failed to check existing subscription")
	}
	if existing != nil {
		return ErrEmailAlreadyExists
	}

	// Set default values
	subscription.Status = StatusPending
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = subscription.CreatedAt

	// Create subscription
	if err := s.store.Create(ctx, subscription); err != nil {
		s.log.Error("failed to create subscription", logger.Error(err))
		return s.wrapError(err, "failed to create subscription")
	}

	return nil
}

// ListSubscriptions returns all subscriptions
func (s *ServiceImpl) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	subscriptions, err := s.store.List(ctx)
	if err != nil {
		s.log.Error("failed to list subscriptions", logger.Error(err))
		return nil, s.wrapError(err, "failed to list subscriptions")
	}

	return subscriptions, nil
}

// GetSubscription returns a subscription by ID
func (s *ServiceImpl) GetSubscription(ctx context.Context, id int64) (*Subscription, error) {
	subscription, err := s.store.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get subscription", logger.Error(err))
		return nil, s.wrapError(err, "failed to get subscription")
	}

	if subscription == nil {
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
		s.log.Error("failed to get subscription by email", logger.Error(err))
		return nil, s.wrapError(err, "failed to get subscription by email")
	}

	if subscription == nil {
		return nil, ErrSubscriptionNotFound
	}

	return subscription, nil
}

// UpdateSubscriptionStatus updates the status of a subscription
func (s *ServiceImpl) UpdateSubscriptionStatus(ctx context.Context, id int64, status Status) error {
	// Validate status
	switch status {
	case StatusPending, StatusActive, StatusCancelled:
		// Valid status
	default:
		return ErrInvalidStatus
	}

	// Check if subscription exists
	subscription, err := s.store.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get subscription", logger.Error(err))
		return s.wrapError(err, "failed to get subscription")
	}

	if subscription == nil {
		return ErrSubscriptionNotFound
	}

	// Update status
	if err := s.store.UpdateStatus(ctx, id, status); err != nil {
		s.log.Error("failed to update subscription status", logger.Error(err))
		return s.wrapError(err, "failed to update subscription status")
	}

	return nil
}

// DeleteSubscription removes a subscription
func (s *ServiceImpl) DeleteSubscription(ctx context.Context, id int64) error {
	// Check if subscription exists
	subscription, err := s.store.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get subscription", logger.Error(err))
		return s.wrapError(err, "failed to get subscription")
	}

	if subscription == nil {
		return ErrSubscriptionNotFound
	}

	// Delete subscription
	if err := s.store.Delete(ctx, id); err != nil {
		s.log.Error("failed to delete subscription", logger.Error(err))
		return s.wrapError(err, "failed to delete subscription")
	}

	return nil
}
