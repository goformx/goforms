package v1

import (
	"go.uber.org/fx"
)

// Module combines all v1 API handlers and their dependencies
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	fx.Provide(
		NewContactAPI,
		NewSubscriptionAPI,
		NewWebHandler,
	),
)
