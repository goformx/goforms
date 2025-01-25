package subscription

import "context"

// Store defines the interface for subscription persistence
type Store interface {
	// Create creates a new subscription
	Create(ctx context.Context, subscription *Subscription) error

	// List returns all subscriptions
	List(ctx context.Context) ([]Subscription, error)

	// Get returns a subscription by ID
	Get(ctx context.Context, id int64) (*Subscription, error)

	// GetByID returns a subscription by ID
	GetByID(ctx context.Context, id int64) (*Subscription, error)

	// GetByEmail returns a subscription by email
	GetByEmail(ctx context.Context, email string) (*Subscription, error)

	// UpdateStatus updates the status of a subscription
	UpdateStatus(ctx context.Context, id int64, status Status) error

	// Delete removes a subscription
	Delete(ctx context.Context, id int64) error
}
