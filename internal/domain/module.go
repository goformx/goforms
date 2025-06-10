// Package domain provides domain services and their dependency injection setup.
// This module is responsible for providing domain services and interfaces,
// while keeping implementation details in the infrastructure layer.
package domain

import (
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/services/auth"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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

	Store          form.Repository
	EventPublisher event.Publisher
	Logger         logging.Logger
}

// NewFormService creates a new form service with dependencies
func NewFormService(p FormServiceParams) (form.Service, error) {
	if p.Store == nil {
		return nil, errors.New("form repository is required")
	}
	if p.EventPublisher == nil {
		return nil, errors.New("event publisher is required")
	}
	if p.Logger == nil {
		return nil, errors.New("logger is required")
	}
	return form.NewService(p.Store, p.EventPublisher, p.Logger), nil
}

// AuthServiceParams contains dependencies for creating an auth service
type AuthServiceParams struct {
	fx.In

	UserService user.Service
	Logger      logging.Logger
}

// NewAuthService creates a new auth service with dependencies
func NewAuthService(p AuthServiceParams) (auth.Service, error) {
	if p.UserService == nil {
		return nil, errors.New("user service is required")
	}
	if p.Logger == nil {
		return nil, errors.New("logger is required")
	}
	return auth.NewService(p.UserService, p.Logger), nil
}

// Module provides all domain services and interfaces
var Module = fx.Options(
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
		// Auth service
		fx.Annotate(
			NewAuthService,
			fx.As(new(auth.Service)),
		),
	),
)

// DomainModule provides domain dependencies
var DomainModule = fx.Options(
	fx.Provide(
		func(
			userRepo user.Repository,
			formRepo form.Repository,
			publisher event.Publisher,
			logger logging.Logger,
		) (user.Service, form.Service) {
			return user.NewService(userRepo, logger), form.NewService(formRepo, publisher, logger)
		},
	),
)
