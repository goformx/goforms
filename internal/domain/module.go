package domain

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// UserServiceParams contains dependencies for creating a user service
type UserServiceParams struct {
	fx.In

	Store  user.Store
	Logger logging.Logger
	Config *config.Config
}

// NewUserService creates a new user service with dependencies
func NewUserService(p UserServiceParams) user.Service {
	return user.NewService(p.Store, p.Logger, p.Config.Security.JWTSecret)
}

// FormServiceParams contains dependencies for creating a form service
type FormServiceParams struct {
	fx.In

	Store form.Store
}

// NewFormService creates a new form service with dependencies
func NewFormService(p FormServiceParams) form.Service {
	return form.NewService(p.Store)
}

// Module combines all domain services
var Module = fx.Options(
	fx.Provide(
		// Contact service
		fx.Annotate(
			contact.NewService,
			fx.As(new(contact.Service)),
		),
		// Subscription service
		fx.Annotate(
			subscription.NewService,
			fx.As(new(subscription.Service)),
		),
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
