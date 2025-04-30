package infrastructure

import (
	"fmt"

	"go.uber.org/fx"

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

// HandlerParams contains dependencies for creating handlers.
// This struct is used with fx.In to inject dependencies into handlers.
// Each field represents a required dependency that must be provided
// by the fx container.
type HandlerParams struct {
	fx.In

	Logger              logging.Logger
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

// NoopHandler is a handler that does nothing
type NoopHandler struct{}

// Register implements the Handler interface
func (h *NoopHandler) Register(e *echo.Echo) {}

// Module combines all infrastructure-level modules and providers
var Module = fx.Options(
	// Core infrastructure
	fx.Provide(
		config.New,
		database.NewDB,
	),

	// Stores
	fx.Provide(
		NewStores,
	),

	// Presentation
	fx.Provide(
		server.New,
	),

	// Handlers
	fx.Provide(
		AsHandler(func(p HandlerParams) *wh.HomeHandler {
			return wh.NewHomeHandler(p.Logger, p.Renderer)
		}),
		AsHandler(func(p HandlerParams) *wh.DemoHandler {
			return wh.NewDemoHandler(p.Logger, p.Renderer, p.SubscriptionService)
		}),
		AsHandler(func(p HandlerParams) *ah.DashboardHandler {
			return ah.NewDashboardHandler(p.Logger, p.Renderer, p.UserService, p.FormService)
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
		FormStore:         formstore.NewStore(db, logger),
	}

	logger.Debug("database stores created",
		logging.Bool("contact_store_available", stores.ContactStore != nil),
		logging.Bool("subscription_store_available", stores.SubscriptionStore != nil),
		logging.Bool("user_store_available", stores.UserStore != nil),
		logging.Bool("form_store_available", stores.FormStore != nil),
	)

	return stores
}
