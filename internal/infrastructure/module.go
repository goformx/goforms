package infrastructure

import (
	"context"

	"go.uber.org/fx"

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
		// Provide logger with config values
		fx.Annotate(
			func(cfg *config.Config) logging.Logger {
				return logging.NewLogger(cfg.App.Debug, cfg.App.Name)
			},
			fx.As(new(logging.Logger)),
		),
		database.NewDB,
	),

	// Infrastructure
	fx.Provide(
		NewStores,
		server.New,
		http.NewHandlers,
	),

	// Lifecycle hooks
	fx.Invoke(
		registerHandlers,
		registerDatabaseHooks,
	),
)

func registerHandlers(srv *server.Server, handlers *http.Handlers) {
	handlers.Register(srv.Echo())
}

func registerDatabaseHooks(lc fx.Lifecycle, db *database.Database, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database connection")
			return db.Close()
		},
	})
}

// Stores groups all database store providers
type Stores struct {
	fx.Out

	ContactStore      contact.Store
	SubscriptionStore subscription.Store
	UserStore         user.Store
}

// NewStores creates all database stores
func NewStores(db *database.Database, logger logging.Logger) Stores {
	return Stores{
		ContactStore:      store.NewContactStore(db, logger),
		SubscriptionStore: store.NewSubscriptionStore(db, logger),
		UserStore:         store.NewUserStore(db, logger),
	}
}
