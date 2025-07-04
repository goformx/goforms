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
	"fmt"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/domain/common/interfaces"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
)

// UserServiceParams contains dependencies for creating a user service
type UserServiceParams struct {
	fx.In

	Repo   user.Repository
	Logger interfaces.Logger
}

// NewUserService creates a new user service with dependencies
func NewUserService(p UserServiceParams) (user.Service, error) {
	fmt.Printf("[DEBUG] NewUserService called with Repo: %T, Logger: %T\n", p.Repo, p.Logger)

	if p.Repo == nil {
		fmt.Println("[DEBUG] NewUserService: Repo is nil!")

		return nil, errors.New("user repository is required")
	}

	if p.Logger == nil {
		fmt.Println("[DEBUG] NewUserService: Logger is nil!")

		return nil, errors.New("logger is required")
	}

	return user.NewService(p.Repo, p.Logger), nil
}

// FormServiceParams contains dependencies for creating a form service
type FormServiceParams struct {
	fx.In

	Repository form.Repository
	EventBus   events.EventBus
	Logger     interfaces.Logger
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
	),
)
