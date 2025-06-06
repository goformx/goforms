package persistence

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	formstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/form"
	userstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/user"
)

// Module provides persistence dependencies
var Module = fx.Module("persistence",
	// Database
	fx.Provide(
		database.NewDB,
	),

	// Stores
	fx.Provide(
		fx.Annotate(
			userstore.NewStore,
			fx.As(new(user.Store)),
		),
		fx.Annotate(
			formstore.NewStore,
			fx.As(new(form.Store)),
		),
	),
)

// StoreParams and NewStores have been removed as they are no longer needed.
