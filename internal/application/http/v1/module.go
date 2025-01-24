package v1

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

// Module combines all v1 API handlers and their dependencies
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	fx.Provide(
		NewContactAPI,
		NewSubscriptionAPI,
		NewWebHandler,
		contact.NewService,
		subscription.NewService,
	),
)
