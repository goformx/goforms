package persistence

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	contactstore "github.com/jonesrussell/goforms/internal/infrastructure/persistence/store/contact"
	subscriptionstore "github.com/jonesrussell/goforms/internal/infrastructure/persistence/store/subscription"
	userstore "github.com/jonesrussell/goforms/internal/infrastructure/persistence/store/user"
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

	DB     *database.Database
	Logger logging.Logger
}

// NewStores creates all database stores
func NewStores(p StoreParams) (contact.Store, subscription.Store, user.Store, error) {
	p.Logger.Debug("creating database stores",
		logging.Bool("db_available", p.DB != nil),
	)

	contactStore := contactstore.NewStore(p.DB, p.Logger)
	subscriptionStore := subscriptionstore.NewStore(p.DB, p.Logger)
	userStore := userstore.NewStore(p.DB, p.Logger)

	return contactStore, subscriptionStore, userStore, nil
}
