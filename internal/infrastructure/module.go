package infrastructure

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence/database"
)

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	fx.Provide(
		config.New,
		database.New,
		database.NewContactStore,
		database.NewSubscriptionStore,
		logging.GetLogger,
	),
)
