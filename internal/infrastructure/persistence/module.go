package persistence

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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
		NewStores,
		fx.Annotate(
			userstore.NewStore,
			fx.As(new(user.Store)),
		),
	),
)

// StoreParams contains dependencies for creating stores
type StoreParams struct {
	fx.In

	DB     *database.Database
	Logger logging.Logger
}

// NewStores creates and returns all required stores
func NewStores(p StoreParams) (
	userStore user.Store,
	err error,
) {
	p.Logger.Debug("creating database stores",
		logging.BoolField("db_available", p.DB != nil),
	)

	userStore = userstore.NewStore(p.DB, p.Logger)

	return userStore, nil
}
