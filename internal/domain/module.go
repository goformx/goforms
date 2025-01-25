package domain

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
)

// Module combines all domain services
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			contact.NewService,
			fx.As(new(contact.Service)),
		),
		fx.Annotate(
			subscription.NewService,
			fx.As(new(subscription.Service)),
		),
		fx.Annotate(
			user.NewService,
			fx.As(new(user.Service)),
		),
	),
)
