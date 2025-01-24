package infrastructure

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/store"
)

// Stores groups all database store providers for better organization
type Stores struct {
	fx.Out

	ContactStore      contact.Store
	SubscriptionStore subscription.Store
	UserStore         user.Store
}

// NewStores creates all database stores
func NewStores(db *database.DB, log logging.Logger) Stores {
	return Stores{
		ContactStore:      database.NewContactStore(db, log),
		SubscriptionStore: database.NewSubscriptionStore(db, log),
		UserStore:         store.NewStore(db.DB, log),
	}
}

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	fx.Provide(
		config.New,
		database.New,
		NewStores, // Consolidated store providers
	),
)
