// Package user provides user repository interfaces for domain persistence.
//
//go:generate mockgen -typed -source=repository.go -destination=../../../test/mocks/user/mock_repository.go -package=user
package user

import (
	"context"

	"github.com/goformx/goforms/internal/domain/common/repository"
	"github.com/goformx/goforms/internal/domain/entities"
)

// Repository defines the interface for user storage
type Repository interface {
	repository.Repository[*entities.User]
	// GetByEmail gets a user by email
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	// GetByUsername gets a user by username
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	// GetByRole gets users by role
	GetByRole(ctx context.Context, role string, offset, limit int) ([]*entities.User, error)
	// GetActiveUsers gets all active users
	GetActiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, error)
	// GetInactiveUsers gets all inactive users
	GetInactiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, error)
}
