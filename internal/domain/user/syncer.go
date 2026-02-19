// Package user provides user domain types and interfaces.
package user

import "context"

// LaravelUserSyncer ensures a Go user row exists for a Laravel user ID (for forms FK).
// Used when handling assertion-authenticated requests so forms.user_id can reference users.uuid.
type LaravelUserSyncer interface {
	// EnsureUser ensures a user row exists with the given ID; creates a shadow user if not.
	EnsureUser(ctx context.Context, userID string) error
}
