package core

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/jonesrussell/goforms/internal/core/subscription"
)

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	fx.Provide(
		contact.NewService,
		subscription.NewService,
	),
)
