// Package repository provides the user repository implementation.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
)

// Syncer ensures a Go user row exists for a Laravel user ID (lazy sync for forms FK).
type Syncer struct {
	repo user.Repository
}

// NewLaravelUserSyncer returns a LaravelUserSyncer that uses the given user repository.
func NewLaravelUserSyncer(repo user.Repository) user.LaravelUserSyncer {
	return &Syncer{repo: repo}
}

// EnsureUser ensures a user row exists with the given ID; creates a shadow user if not.
func (s *Syncer) EnsureUser(ctx context.Context, userID string) error {
	_, err := s.repo.GetByID(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, common.ErrNotFound) {
		return fmt.Errorf("get user by ID: %w", err)
	}
	shadow := entities.NewLaravelShadowUser(userID)
	if createErr := s.repo.Create(ctx, shadow); createErr != nil {
		return fmt.Errorf("create Laravel shadow user: %w", createErr)
	}
	return nil
}
