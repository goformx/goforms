package persistence

import (
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence/database"
	contactstore "github.com/jonesrussell/goforms/internal/infrastructure/persistence/store/contact"
	subscriptionstore "github.com/jonesrussell/goforms/internal/infrastructure/persistence/store/subscription"
	userstore "github.com/jonesrussell/goforms/internal/infrastructure/persistence/store/user"
	"go.uber.org/fx"
)

// Module provides persistence dependencies
var Module = fx.Module("persistence",
	// Database
	fx.Provide(
		database.NewConfig,
		database.NewDB,
	),

	// Stores
	fx.Provide(
		fx.Annotate(
			contactstore.NewStore,
			fx.As(new(contact.Store)),
		),
		fx.Annotate(
			subscriptionstore.NewStore,
			fx.As(new(subscription.Store)),
		),
		fx.Annotate(
			userstore.NewStore,
			fx.As(new(user.Store)),
		),
	),
)

// StoreParams contains dependencies for creating stores
type StoreParams struct {
	fx.In

	DB     *database.DB
	Logger logging.Logger
}

// NewStores creates all database stores
func NewStores(p StoreParams) error {
	p.Logger.Debug("creating database stores",
		logging.Bool("db_available", p.DB != nil),
	)
	return nil
}
