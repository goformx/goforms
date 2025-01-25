package infrastructure

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/store"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

// Module combines all infrastructure-level modules and providers
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	// Core infrastructure
	fx.Provide(
		config.New,
		fx.Annotate(
			logging.NewLogger,
			fx.ParamTags(`name:"debug"`, `name:"app_name"`),
		),
		// Database setup
		fx.Annotate(
			func(cfg *config.Config) database.Config {
				return database.Config{
					Host:     cfg.Database.Host,
					Port:     cfg.Database.Port,
					User:     cfg.Database.User,
					Password: cfg.Database.Password,
					Database: cfg.Database.Name,
				}
			},
		),
		database.New,
		store.NewStore,
	),

	// Handlers
	fx.Provide(
		fx.Annotate(
			func(p HandlerParams) []handler.Handler {
				base := handler.Base{Logger: p.Logger}
				return []handler.Handler{
					handler.NewVersionHandler(p.VersionInfo, base),
					handler.NewWebHandler(p.Logger, p.ContactService, p.Renderer),
				}
			},
			fx.ResultTags(`group:"handlers"`),
		),
	),

	// Logger dependencies
	fx.Provide(
		fx.Annotate(
			func(cfg *config.Config) bool {
				return cfg.App.Debug
			},
			fx.ResultTags(`name:"debug"`),
		),
		fx.Annotate(
			func(cfg *config.Config) string {
				return cfg.App.Name
			},
			fx.ResultTags(`name:"app_name"`),
		),
	),

	// Lifecycle hooks
	fx.Invoke(
		registerDatabaseHooks,
	),
)

// HandlerParams contains dependencies for creating handlers
type HandlerParams struct {
	fx.In

	Logger         logging.Logger
	VersionInfo    handler.VersionInfo `name:"version_info"`
	Renderer       *view.Renderer
	ContactService contact.Service
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

func registerDatabaseHooks(lc fx.Lifecycle, db *sqlx.DB, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database connection")
			return db.Close()
		},
	})
}
