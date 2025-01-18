package api

import (
	"go.uber.org/fx"

	v1 "github.com/jonesrussell/goforms/internal/api/v1"
)

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	fx.Provide(
		v1.NewContactAPI,
		v1.NewSubscriptionAPI,
	),
)
