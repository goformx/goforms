// Package domain provides domain services and their dependency injection setup.
// This module is responsible for providing domain services and interfaces,
// while keeping implementation details in the infrastructure layer.
//
// The domain layer follows clean architecture principles:
// - Entities: Core business objects
// - Services: Business logic and use cases
// - Repositories: Data access interfaces
// - Events: Domain events for cross-cutting concerns
package domain

import (
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	formstore "github.com/goformx/goforms/internal/infrastructure/repository/form"
	formsubmissionstore "github.com/goformx/goforms/internal/infrastructure/repository/form/submission"
	userstore "github.com/goformx/goforms/internal/infrastructure/repository/user"
)

// UserServiceParams contains dependencies for creating a user service
type UserServiceParams struct {
	fx.In

	Repo   user.Repository
	Logger logging.Logger
}

// NewUserService creates a new user service with dependencies
func NewUserService(p UserServiceParams) (user.Service, error) {
	if p.Repo == nil {
		return nil, errors.New("user repository is required")
	}

	if p.Logger == nil {
		return nil, errors.New("logger is required")
	}

	return user.NewService(p.Repo, p.Logger), nil
}

// FormServiceParams contains dependencies for creating a form service
type FormServiceParams struct {
	fx.In

	Repository form.Repository
	EventBus   events.EventBus
	Logger     logging.Logger
}

// NewFormService creates a new form service with dependencies
func NewFormService(p FormServiceParams) (form.Service, error) {
	if p.Repository == nil {
		return nil, errors.New("form repository is required")
	}

	if p.EventBus == nil {
		return nil, errors.New("event bus is required")
	}

	if p.Logger == nil {
		return nil, errors.New("logger is required")
	}

	return form.NewService(p.Repository, p.EventBus, p.Logger), nil
}

// StoreParams groups store dependencies
type StoreParams struct {
	fx.In
	DB     database.DB
	Logger logging.Logger
}

// Stores groups all store implementations
type Stores struct {
	fx.Out
	UserRepository           user.Repository
	FormRepository           form.Repository
	FormSubmissionRepository form.SubmissionRepository
}

// NewStores creates new store instances with proper validation and error handling
func NewStores(p StoreParams) (Stores, error) {
	if p.DB == nil {
		return Stores{}, errors.New("database connection is required")
	}

	if p.Logger == nil {
		return Stores{}, errors.New("logger is required")
	}

	// Initialize repositories using the interface
	userRepo := userstore.NewStore(p.DB, p.Logger)
	formRepo := formstore.NewStore(p.DB, p.Logger)
	formSubmissionRepo := formsubmissionstore.NewStore(p.DB, p.Logger)

	// Validate repository instances
	if userRepo == nil || formRepo == nil || formSubmissionRepo == nil {
		p.Logger.Error("failed to create repository",
			"operation", "repository_initialization",
			"repository_type", "user/form/submission",
			"error_type", "nil_repository",
		)

		return Stores{}, errors.New("failed to create repository: one or more repositories are nil")
	}

	return Stores{
		Out:                      fx.Out{},
		UserRepository:           userRepo,
		FormRepository:           formRepo,
		FormSubmissionRepository: formSubmissionRepo,
	}, nil
}

// Module provides all domain layer dependencies
var Module = fx.Module("domain",
	fx.Provide(
		// User service
		fx.Annotate(
			NewUserService,
			fx.As(new(user.Service)),
		),
		// Form service
		fx.Annotate(
			NewFormService,
			fx.As(new(form.Service)),
		),
		NewStores,
		// User ensurer (ensures Go user row exists for assertion-authenticated requests)
		fx.Annotate(
			userstore.NewUserEnsurer,
			fx.As(new(user.UserEnsurer)),
		),
	),
)
