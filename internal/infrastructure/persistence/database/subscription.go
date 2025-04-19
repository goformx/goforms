package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// SubscriptionStore implements subscription.Store
type SubscriptionStore struct {
	db     *sqlx.DB
	logger logging.Logger
}

// NewSubscriptionStore creates a new subscription store
func NewSubscriptionStore(db *sqlx.DB, logger logging.Logger) subscription.Store {
	return &SubscriptionStore{
		db:     db,
		logger: logger,
	}
}

// Create creates a new subscription
func (s *SubscriptionStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	query := `
		INSERT INTO subscriptions (email, name, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		sub.Email,
		sub.Name,
		sub.Status,
		sub.CreatedAt,
		sub.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	sub.ID = id
	return nil
}

// List returns all subscriptions
func (s *SubscriptionStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	var subs []subscription.Subscription
	query := `
		SELECT id, email, name, status, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
	`

	if err := s.db.SelectContext(ctx, &subs, query); err != nil {
		return nil, err
	}

	return subs, nil
}

// Get implements subscription.Store
func (s *SubscriptionStore) Get(ctx context.Context, id int64) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	query := `SELECT * FROM subscriptions WHERE id = ?`

	if err := s.db.GetContext(ctx, &sub, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, subscription.ErrSubscriptionNotFound
		}
		s.logger.Error("failed to get subscription", logging.Error(err))
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

// GetByID implements subscription.Store
func (s *SubscriptionStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	return s.Get(ctx, id)
}

// GetByEmail returns a subscription by email
func (s *SubscriptionStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	query := `
		SELECT id, email, name, status, created_at, updated_at
		FROM subscriptions
		WHERE email = ?
	`

	err := s.db.GetContext(ctx, &sub, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

// UpdateStatus updates the status of a subscription
func (s *SubscriptionStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	query := `
		UPDATE subscriptions
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return subscription.ErrSubscriptionNotFound
	}

	return nil
}

// Delete removes a subscription
func (s *SubscriptionStore) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM subscriptions
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return subscription.ErrSubscriptionNotFound
	}

	return nil
}
