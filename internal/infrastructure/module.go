package infrastructure

import (
	"fmt"

	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/store"
	"github.com/jonesrussell/goforms/internal/presentation/handlers"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

// AsHandler annotates the given constructor to state that
// it provides a handler to the "handlers" group.
// This is used to register handlers with the fx dependency injection container.
// Each handler must be annotated with this function to be properly registered.
//
// Example:
//
//	AsHandler(func(logger logging.Logger, svc SomeService) *handler.SomeHandler {
//	    return handler.NewSomeHandler(logger, handler.WithSomeService(svc))
//	})
func AsHandler(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(handler.Handler)),
		fx.ResultTags(`group:"handlers"`),
	)
}

// HandlerParams contains dependencies for creating handlers.
// This struct is used with fx.In to inject dependencies into handlers.
// Each field represents a required dependency that must be provided
// by the fx container.
type HandlerParams struct {
	fx.In

	Logger              logging.Logger
	VersionInfo         handler.VersionInfo
	Renderer            *view.Renderer
	ContactService      contact.Service
	SubscriptionService subscription.Service
	UserService         user.Service
	FormService         form.Service
	Config              *config.Config
}

// Stores groups all database store providers.
// This struct is used with fx.Out to provide multiple stores
// to the fx container in a single provider function.
type Stores struct {
	fx.Out

	ContactStore      contact.Store
	SubscriptionStore subscription.Store
	UserStore         user.Store
	FormStore         form.Store
}

// Module combines all infrastructure-level modules and providers
var Module = fx.Module("infrastructure",
	fx.Provide(
		config.New,
		func(cfg *config.Config) (logging.Logger, error) {
			return logging.NewLogger(cfg.App.Debug, cfg.App.Name)
		},
		database.NewDB,
		persistence.NewStores,
		server.New,
		// Services
		fx.Annotate(
			contact.NewService,
			fx.As(new(contact.Service)),
		),
		fx.Annotate(
			subscription.NewService,
			fx.As(new(subscription.Service)),
		),
		fx.Annotate(
			user.NewService,
			fx.As(new(user.Service)),
		),
		fx.Annotate(
			form.NewService,
			fx.As(new(form.Service)),
		),
		// Handlers
		AsHandler(func(p HandlerParams) *handler.WebHandler {
			return handler.NewWebHandler(p.Logger,
				handler.WithRenderer(p.Renderer),
				handler.WithContactService(p.ContactService),
				handler.WithWebSubscriptionService(p.SubscriptionService),
				handler.WithWebDebug(p.Config.App.Debug),
			)
		}),
		AsHandler(func(p HandlerParams) *handler.AuthHandler {
			return handler.NewAuthHandler(p.Logger,
				handler.WithUserService(p.UserService),
			)
		}),
		AsHandler(func(p HandlerParams) *handler.ContactHandler {
			return handler.NewContactHandler(p.Logger,
				handler.WithContactServiceOpt(p.ContactService),
			)
		}),
		AsHandler(func(p HandlerParams) *handler.SubscriptionHandler {
			return handler.NewSubscriptionHandler(p.Logger,
				handler.WithSubscriptionService(p.SubscriptionService),
			)
		}),
		AsHandler(func(p HandlerParams) *handlers.DashboardHandler {
			return handlers.NewDashboardHandler(p.UserService, p.FormService)
		}),
	),
)

// NewStores creates all database stores.
// This function is responsible for initializing all database stores
// and providing them to the fx container.
func NewStores(db *database.Database, logger logging.Logger) Stores {
	logger.Debug("creating database stores",
		logging.Bool("database_available", db != nil),
		logging.String("database_type", fmt.Sprintf("%T", db)),
	)

	stores := Stores{
		ContactStore:      store.NewContactStore(db, logger),
		SubscriptionStore: store.NewSubscriptionStore(db, logger),
		UserStore:         store.NewUserStore(db, logger),
		FormStore:         store.NewFormStore(db, logger),
	}

	logger.Debug("database stores created",
		logging.Bool("contact_store_available", stores.ContactStore != nil),
		logging.Bool("subscription_store_available", stores.SubscriptionStore != nil),
		logging.Bool("user_store_available", stores.UserStore != nil),
		logging.Bool("form_store_available", stores.FormStore != nil),
	)

	return stores
}

// NewHandlers creates all application handlers
func NewHandlers(params HandlerParams) []handler.Handler {
	routeHandlers := []handler.Handler{
		handler.NewAuthHandler(
			params.Logger,
			handler.WithUserService(params.UserService),
		),
		handler.NewWebHandler(
			params.Logger,
			handler.WithRenderer(params.Renderer),
			handler.WithContactService(params.ContactService),
			handler.WithWebSubscriptionService(params.SubscriptionService),
			handler.WithWebDebug(params.Config.App.Debug),
		),
		handlers.NewDashboardHandler(
			params.UserService,
			params.FormService,
		),
	}

	return routeHandlers
}
