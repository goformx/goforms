package domain

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// UserServiceParams contains dependencies for creating a user service
type UserServiceParams struct {
	fx.In

	Repo   user.Repository
	Logger logging.Logger
}

// NewUserService creates a new user service with dependencies
func NewUserService(p UserServiceParams) user.Service {
	return user.NewService(p.Repo, p.Logger)
}

// FormServiceParams contains dependencies for creating a form service
type FormServiceParams struct {
	fx.In

	Store          form.Repository
	EventPublisher event.Publisher
	Logger         logging.Logger
}

// NewFormService creates a new form service with dependencies
func NewFormService(p FormServiceParams) form.Service {
	return form.NewService(p.Store, p.EventPublisher, p.Logger)
}

// Module combines all domain services
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
		// Event publisher
		fx.Annotate(
			event.NewMemoryPublisher,
			fx.As(new(event.Publisher)),
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
