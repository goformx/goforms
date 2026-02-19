// Package user provides user domain types and interfaces.
package user

import "context"

// UserEnsurer ensures a user row exists for a given user ID (for forms FK).
// Called before form operations so forms.user_id can reference users.uuid.
type UserEnsurer interface {
	// EnsureUser ensures a user row exists with the given ID; creates a shadow user if not.
	EnsureUser(ctx context.Context, userID string) error
}
