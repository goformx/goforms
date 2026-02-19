package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
)

const (
	// shadowPassword is an invalid bcrypt hash that can never match any password.
	// This ensures shadow users cannot log in via Go.
	shadowPassword = "!shadow-no-login"

	// shadowEmailDomain uses a dotted domain so it passes isValidEmail checks.
	shadowEmailDomain = "@shadow.local"

	// maxEmailLength matches the DB column size constraint.
	maxEmailLength = 255
)

// Syncer ensures a Go user row exists for a given user ID (lazy sync for forms FK).
type Syncer struct {
	repo user.Repository
}

// NewUserEnsurer returns a UserEnsurer that uses the given user repository.
func NewUserEnsurer(repo user.Repository) user.UserEnsurer {
	return &Syncer{repo: repo}
}

// EnsureUser checks via Repository.GetByID; on ErrNotFound, creates a shadow user with retry on race.
func (s *Syncer) EnsureUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("ensure user: user ID must not be empty")
	}
	_, err := s.repo.GetByID(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, common.ErrNotFound) {
		return fmt.Errorf("get user by ID: %w", err)
	}
	shadow := newShadowUser(userID)
	if createErr := s.repo.Create(ctx, shadow); createErr != nil {
		// Another request may have created the user concurrently; verify before failing.
		if _, retryErr := s.repo.GetByID(ctx, userID); retryErr == nil {
			return nil
		}
		return fmt.Errorf("create shadow user: %w", createErr)
	}
	return nil
}

// newShadowUser returns a minimal user so forms.user_id FK is satisfied.
// The user is not intended for login via Go (placeholder email and invalid password hash).
func newShadowUser(id string) *entities.User {
	email := "laravel-" + id + shadowEmailDomain
	if len(email) > maxEmailLength {
		email = email[:maxEmailLength]
	}
	return &entities.User{
		ID:             id,
		Email:          email,
		HashedPassword: shadowPassword,
		FirstName:      "Laravel",
		LastName:       "Sync",
		Role:           "user",
		Active:         true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      gorm.DeletedAt{},
	}
}
