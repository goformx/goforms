package infrastructure

import (
	"context"
	"fmt"

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

// AsHandler annotates the given constructor to state that
// it provides a handler to the "handlers" group
func AsHandler(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(handler.Handler)),
		fx.ResultTags(`group:"handlers"`),
	)
}

// Module combines all infrastructure-level modules and providers
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	// Core infrastructure
	fx.Provide(
		// Config must be provided first
		config.New,

		// Logger setup
		func(cfg *config.Config) bool {
			return cfg.App.Debug
		},
		func(cfg *config.Config) string {
			return cfg.App.Name
		},
		logging.NewLogger,

		// Database setup
		func(cfg *config.Config, logger logging.Logger) (*database.Database, error) {
			logger.Debug("initializing database",
				logging.String("host", cfg.Database.Host),
				logging.Int("port", cfg.Database.Port),
				logging.String("name", cfg.Database.Name),
				logging.String("user", cfg.Database.User),
			)
			return database.NewDB(cfg, logger)
		},
		NewStores,

		// Handlers
		AsHandler(handler.NewWebHandler),
		AsHandler(handler.NewAuthHandler),
		AsHandler(handler.NewContactHandler),
		AsHandler(handler.NewSubscriptionHandler),
	),

	// Lifecycle hooks
	fx.Invoke(
		registerDatabaseHooks,
	),
)

// HandlerParams contains dependencies for creating handlers
type HandlerParams struct {
	fx.In

	Logger              logging.Logger
	VersionInfo         handler.VersionInfo
	Renderer            *view.Renderer
	ContactService      contact.Service
	SubscriptionService subscription.Service
	UserService         user.Service
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
	logger.Debug("creating database stores",
		logging.Bool("database_available", db != nil),
		logging.String("database_type", fmt.Sprintf("%T", db)),
	)

	stores := Stores{
		ContactStore:      store.NewContactStore(db, logger),
		SubscriptionStore: store.NewSubscriptionStore(db, logger),
		UserStore:         store.NewUserStore(db, logger),
	}

	logger.Debug("database stores created",
		logging.Bool("contact_store_available", stores.ContactStore != nil),
		logging.Bool("subscription_store_available", stores.SubscriptionStore != nil),
		logging.Bool("user_store_available", stores.UserStore != nil),
	)

	return stores
}

func registerDatabaseHooks(lc fx.Lifecycle, db *database.Database, logger logging.Logger) {
	logger.Debug("registering database lifecycle hooks",
		logging.Bool("database_available", db != nil),
		logging.Bool("lifecycle_available", lc != nil),
	)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Debug("database starting")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database connection")
			if err := db.Close(); err != nil {
				logger.Error("failed to close database connection", logging.Error(err))
				return fmt.Errorf("failed to close database connection: %w", err)
			}
			logger.Debug("database connection closed successfully")
			return nil
		},
	})

	logger.Debug("database lifecycle hooks registered successfully")
}

// NewHandlers creates all application handlers
func NewHandlers(p HandlerParams) []handler.Handler {
	p.Logger.Debug("creating handlers",
		logging.String("version", p.VersionInfo.Version),
		logging.Bool("renderer_available", p.Renderer != nil),
		logging.Bool("contact_service_available", p.ContactService != nil),
		logging.Bool("subscription_service_available", p.SubscriptionService != nil),
		logging.Bool("user_service_available", p.UserService != nil),
	)

	p.Logger.Debug("creating web handler")
	webHandler := handler.NewWebHandler(p.Logger, p.ContactService, p.Renderer)
	p.Logger.Debug("web handler created", logging.Bool("handler_available", webHandler != nil))

	p.Logger.Debug("creating auth handler")
	authHandler := handler.NewAuthHandler(p.Logger, p.UserService)
	p.Logger.Debug("auth handler created", logging.Bool("handler_available", authHandler != nil))

	p.Logger.Debug("creating contact handler")
	contactHandler := handler.NewContactHandler(p.Logger, p.ContactService)
	p.Logger.Debug("contact handler created", logging.Bool("handler_available", contactHandler != nil))

	p.Logger.Debug("creating subscription handler")
	subscriptionHandler := handler.NewSubscriptionHandler(p.Logger, p.SubscriptionService)
	p.Logger.Debug("subscription handler created", logging.Bool("handler_available", subscriptionHandler != nil))

	handlers := []handler.Handler{
		webHandler,
		authHandler,
		contactHandler,
		subscriptionHandler,
	}

	for i, h := range handlers {
		p.Logger.Debug("registered handler",
			logging.Int("index", i),
			logging.String("type", fmt.Sprintf("%T", h)),
		)
	}

	p.Logger.Debug("all handlers created", logging.Int("count", len(handlers)))
	return handlers
}
