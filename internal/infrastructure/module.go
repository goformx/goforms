package infrastructure

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/http"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/store"
)

// Module exports infrastructure components
var Module = fx.Options(
	// Core dependencies
	fx.Provide(
		config.New,
		logging.NewLogger,
		database.NewDB,
	),

	// Domain services
	domain.Module,

	// Infrastructure
	fx.Provide(
		NewStores,
		server.New,
		http.NewHandlers,
	),

	// Start the server and register routes
	fx.Invoke(func(srv *server.Server, handlers *http.Handlers) {
		handlers.Register(srv.Echo())
	}),
)

// Stores groups all database store providers
type Stores struct {
	fx.Out

	ContactStore      contact.Store      `group:"stores"`
	SubscriptionStore subscription.Store `group:"stores"`
	UserStore         user.Store         `group:"stores"`
}

// NewStores creates all database stores
func NewStores(db *database.Database, logger logging.Logger) Stores {
	return Stores{
		ContactStore:      store.NewContactStore(db, logger),
		SubscriptionStore: store.NewSubscriptionStore(db, logger),
		UserStore:         store.NewUserStore(db, logger),
	}
}
