package infrastructure

import (
	"fmt"

	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	h "github.com/jonesrussell/goforms/internal/handlers"
	wh "github.com/jonesrussell/goforms/internal/handlers/web"
	ah "github.com/jonesrussell/goforms/internal/handlers/web/admin"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/store"
	formstore "github.com/jonesrussell/goforms/internal/infrastructure/store/form"
	"github.com/jonesrussell/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// AsHandler marks a provider as a handler
func AsHandler(fn any) fx.Option {
	return fx.Provide(fx.Annotate(
		fn,
		fx.As(new(h.Handler)),
		fx.ResultTags(`group:"handlers"`),
	))
}

// CoreParams contains core infrastructure dependencies that are commonly needed by handlers.
// These are typically required for basic handler functionality like logging and rendering.
type CoreParams struct {
	fx.In
	Logger   logging.Logger
	Renderer *view.Renderer
	Config   *config.Config
}

// ServiceParams contains all service dependencies that handlers might need.
// This separation makes it easier to manage service dependencies and allows for
// more granular dependency injection.
type ServiceParams struct {
	fx.In
	ContactService      contact.Service
	SubscriptionService subscription.Service
	UserService         user.Service
	FormService         form.Service
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

// NoopHandler is a handler that does nothing
type NoopHandler struct{}

// Register implements the Handler interface
func (nh *NoopHandler) Register(e *echo.Echo) {}

// InfrastructureModule provides core infrastructure dependencies.
// This module includes configuration and database setup.
var InfrastructureModule = fx.Options(
	fx.Provide(
		config.New,
		database.NewDB,
	),
)

// StoreModule provides all database store implementations.
// This module is responsible for creating and managing database stores.
var StoreModule = fx.Options(
	fx.Provide(NewStores),
)

// HandlerModule provides all HTTP handlers for the application.
// This module is responsible for setting up route handlers and their dependencies.
var HandlerModule = fx.Options(
	// Web handlers
	AsHandler(func(core CoreParams) *wh.HomeHandler {
		return wh.NewHomeHandler(core.Logger, core.Renderer)
	}),
	AsHandler(func(core CoreParams, services ServiceParams) *wh.DemoHandler {
		return wh.NewDemoHandler(core.Logger, core.Renderer, services.SubscriptionService)
	}),
	AsHandler(func(core CoreParams, services ServiceParams) *ah.DashboardHandler {
		return ah.NewDashboardHandler(core.Logger, core.Renderer, services.UserService, services.FormService)
	}),
	AsHandler(func(core CoreParams, services ServiceParams) (h.Handler, error) {
		handler, err := handler.NewWebHandler(core.Logger,
			handler.WithRenderer(core.Renderer),
			handler.WithContactService(services.ContactService),
			handler.WithWebSubscriptionService(services.SubscriptionService),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create web handler: %w", err)
		}
		return handler, nil
	}),
)

// ServerModule provides the HTTP server setup.
// This module is responsible for creating and configuring the Echo server.
var ServerModule = fx.Options(
	fx.Provide(server.New),
)

// Module combines all infrastructure-level modules into a single module.
// This is the main entry point for infrastructure dependencies.
var Module = fx.Options(
	InfrastructureModule,
	StoreModule,
	ServerModule,
	HandlerModule,
)

// NewStores creates all database stores.
// This function is responsible for initializing all database stores
// and providing them to the fx container.
func NewStores(db *database.Database, logger logging.Logger) (Stores, error) {
	if db == nil {
		return Stores{}, fmt.Errorf("database connection is nil")
	}

	logger.Debug("initializing database stores",
		logging.String("database_type", fmt.Sprintf("%T", db)),
	)

	// Create stores with error handling
	contactStore := store.NewContactStore(db, logger)
	if contactStore == nil {
		return Stores{}, fmt.Errorf("failed to create contact store")
	}

	subscriptionStore := store.NewSubscriptionStore(db, logger)
	if subscriptionStore == nil {
		return Stores{}, fmt.Errorf("failed to create subscription store")
	}

	userStore := store.NewUserStore(db, logger)
	if userStore == nil {
		return Stores{}, fmt.Errorf("failed to create user store")
	}

	formStore := formstore.NewStore(db, logger)
	if formStore == nil {
		return Stores{}, fmt.Errorf("failed to create form store")
	}

	stores := Stores{
		ContactStore:      contactStore,
		SubscriptionStore: subscriptionStore,
		UserStore:         userStore,
		FormStore:         formStore,
	}

	logger.Debug("successfully initialized all database stores",
		logging.String("contact_store_type", fmt.Sprintf("%T", stores.ContactStore)),
		logging.String("subscription_store_type", fmt.Sprintf("%T", stores.SubscriptionStore)),
		logging.String("user_store_type", fmt.Sprintf("%T", stores.UserStore)),
		logging.String("form_store_type", fmt.Sprintf("%T", stores.FormStore)),
	)

	return stores, nil
}
