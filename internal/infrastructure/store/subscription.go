package store

import (
	"context"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

type subscriptionStore struct {
	db     *database.Database
	logger logging.Logger
}

// NewSubscriptionStore creates a new subscription store
func NewSubscriptionStore(db *database.Database, logger logging.Logger) subscription.Store {
	return &subscriptionStore{
		db:     db,
		logger: logger,
	}
}

// Create implements subscription.Store
func (s *subscriptionStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	query := `INSERT INTO subscriptions (email, status, created_at, updated_at)
			  VALUES (:email, :status, NOW(), NOW())`

	_, err := s.db.NamedExecContext(ctx, query, sub)
	if err != nil {
		s.logger.Error("failed to create subscription", logging.Error(err))
		return err
	}

	return nil
}

// List implements subscription.Store
func (s *subscriptionStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	var subs []subscription.Subscription
	query := `SELECT * FROM subscriptions ORDER BY created_at DESC`

	if err := s.db.SelectContext(ctx, &subs, query); err != nil {
		s.logger.Error("failed to list subscriptions", logging.Error(err))
		return nil, err
	}

	return subs, nil
}

// Get implements subscription.Store
func (s *subscriptionStore) Get(ctx context.Context, id int64) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	query := `SELECT * FROM subscriptions WHERE id = ?`

	if err := s.db.GetContext(ctx, &sub, query, id); err != nil {
		s.logger.Error("failed to get subscription", logging.Error(err))
		return nil, err
	}

	return &sub, nil
}

// GetByEmail implements subscription.Store
func (s *subscriptionStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	query := `SELECT * FROM subscriptions WHERE email = ?`

	if err := s.db.GetContext(ctx, &sub, query, email); err != nil {
		s.logger.Error("failed to get subscription by email", logging.Error(err))
		return nil, err
	}

	return &sub, nil
}

// UpdateStatus implements subscription.Store
func (s *subscriptionStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	query := `UPDATE subscriptions SET status = ?, updated_at = NOW() WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		s.logger.Error("failed to update subscription status", logging.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("failed to get rows affected", logging.Error(err))
		return err
	}

	if rows == 0 {
		return subscription.ErrSubscriptionNotFound
	}

	return nil
}

// Delete implements subscription.Store
func (s *subscriptionStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM subscriptions WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		s.logger.Error("failed to delete subscription", logging.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("failed to get rows affected", logging.Error(err))
		return err
	}

	if rows == 0 {
		return subscription.ErrSubscriptionNotFound
	}

	return nil
}

// GetByID implements subscription.Store
func (s *subscriptionStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	return s.Get(ctx, id) // Reuse Get implementation
}

// Other required methods for subscription.Store interface...
