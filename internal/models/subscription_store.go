package models

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrSubscriptionExists = errors.New("email already subscribed")
	ErrDatabaseError      = errors.New("database error")
)

type subscriptionStore struct {
	db *sqlx.DB
}

// NewSubscriptionStore creates a new subscription store
func NewSubscriptionStore(db *sqlx.DB) SubscriptionStore {
	return &subscriptionStore{db: db}
}

// Create creates a new subscription
func (s *subscriptionStore) Create(subscription *Subscription) error {
	// Check for existing subscription
	var exists bool
	err := s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = ?)", subscription.Email)
	if err != nil {
		return ErrDatabaseError
	}
	if exists {
		return ErrSubscriptionExists
	}

	now := time.Now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	query := `
		INSERT INTO subscriptions (email, created_at, updated_at)
		VALUES (?, ?, ?)
		RETURNING id`

	err = s.db.QueryRowx(query,
		subscription.Email,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	).Scan(&subscription.ID)

	if err != nil {
		return ErrDatabaseError
	}

	return nil
}
