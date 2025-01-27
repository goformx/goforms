package subscription

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence/database"
)

// Store implements subscription.Store interface
type Store struct {
	db     *database.DB
	logger logging.Logger
}

// NewStore creates a new subscription store
func NewStore(db *database.DB, logger logging.Logger) subscription.Store {
	logger.Debug("creating subscription store",
		logging.Bool("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create stores a new subscription
func (s *Store) Create(ctx context.Context, sub *subscription.Subscription) error {
	query := `
		INSERT INTO subscriptions (name, email, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	s.logger.Debug("creating subscription",
		logging.String("email", sub.Email),
		logging.String("status", string(sub.Status)),
	)

	err := s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, query,
			sub.Name,
			sub.Email,
			sub.Status,
			sub.CreatedAt,
			sub.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert subscription: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert ID: %w", err)
		}

		sub.ID = id
		return nil
	})

	if err != nil {
		s.logger.Error("failed to create subscription",
			logging.Error(err),
			logging.String("email", sub.Email),
		)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.Info("subscription created",
		logging.Int64("id", sub.ID),
		logging.String("email", sub.Email),
		logging.String("status", string(sub.Status)),
	)

	return nil
}

// List returns all subscriptions
func (s *Store) List(ctx context.Context) ([]subscription.Subscription, error) {
	query := `
		SELECT id, name, email, status, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
	`

	s.logger.Debug("listing subscriptions")

	var subscriptions []subscription.Subscription
	if err := s.db.SelectContext(ctx, &subscriptions, query); err != nil {
		s.logger.Error("failed to list subscriptions",
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	s.logger.Debug("subscriptions retrieved",
		logging.Int("count", len(subscriptions)),
	)

	return subscriptions, nil
}

// Get returns a specific subscription by ID
func (s *Store) Get(ctx context.Context, id int64) (*subscription.Subscription, error) {
	query := `
		SELECT id, name, email, status, created_at, updated_at
		FROM subscriptions
		WHERE id = ?
	`

	s.logger.Debug("getting subscription",
		logging.Int64("id", id),
	)

	var sub subscription.Subscription
	if err := s.db.GetContext(ctx, &sub, query, id); err != nil {
		s.logger.Error("failed to get subscription",
			logging.Error(err),
			logging.Int64("id", id),
		)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	s.logger.Debug("subscription retrieved",
		logging.Int64("id", sub.ID),
		logging.String("email", sub.Email),
		logging.String("status", string(sub.Status)),
	)

	return &sub, nil
}

// GetByID returns a specific subscription by ID
func (s *Store) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	return s.Get(ctx, id)
}

// GetByEmail returns a subscription by email
func (s *Store) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	query := `
		SELECT id, name, email, status, created_at, updated_at
		FROM subscriptions
		WHERE email = ?
	`

	s.logger.Debug("getting subscription by email",
		logging.String("email", email),
	)

	var sub subscription.Subscription
	if err := s.db.GetContext(ctx, &sub, query, email); err != nil {
		s.logger.Error("failed to get subscription by email",
			logging.Error(err),
			logging.String("email", email),
		)
		return nil, fmt.Errorf("failed to get subscription by email: %w", err)
	}

	s.logger.Debug("subscription retrieved",
		logging.Int64("id", sub.ID),
		logging.String("email", sub.Email),
		logging.String("status", string(sub.Status)),
	)

	return &sub, nil
}

// UpdateStatus updates the status of a subscription
func (s *Store) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	query := `
		UPDATE subscriptions
		SET status = ?, updated_at = NOW()
		WHERE id = ?
	`

	s.logger.Debug("updating subscription status",
		logging.Int64("id", id),
		logging.String("status", string(status)),
	)

	err := s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, query, status, id)
		if err != nil {
			return fmt.Errorf("failed to update subscription status: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rows == 0 {
			return fmt.Errorf("subscription not found: %d", id)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to update subscription status",
			logging.Error(err),
			logging.Int64("id", id),
			logging.String("status", string(status)),
		)
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	s.logger.Info("subscription status updated",
		logging.Int64("id", id),
		logging.String("status", string(status)),
	)

	return nil
}

// Delete removes a subscription
func (s *Store) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM subscriptions
		WHERE id = ?
	`

	s.logger.Debug("deleting subscription",
		logging.Int64("id", id),
	)

	err := s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to delete subscription: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rows == 0 {
			return fmt.Errorf("subscription not found: %d", id)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to delete subscription",
			logging.Error(err),
			logging.Int64("id", id),
		)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	s.logger.Info("subscription deleted",
		logging.Int64("id", id),
	)

	return nil
}
