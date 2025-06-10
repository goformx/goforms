// Package infrastructure provides core infrastructure components and their dependency injection setup.
package infrastructure

import (
	"errors"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	formevent "github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/presentation/view"
)

const (
	// MinSecretLength is the minimum length required for security secrets
	MinSecretLength = 32
)

// CoreParams groups core infrastructure dependencies
type CoreParams struct {
	fx.In
	Config   *config.Config
	Logger   logging.Logger
	Renderer view.Renderer
	Echo     *echo.Echo
}

// ServiceParams groups business service dependencies
type ServiceParams struct {
	fx.In
	UserService user.Service
	FormService form.Service
}

// EventPublisherParams contains dependencies for creating an event publisher
type EventPublisherParams struct {
	fx.In
	Logger logging.Logger
}

// NewEventPublisher creates a new event publisher with dependencies
func NewEventPublisher(p EventPublisherParams) (formevent.Publisher, error) {
	if p.Logger == nil {
		return nil, errors.New("logger is required for event publisher")
	}
	return event.NewMemoryPublisher(p.Logger), nil
}

// AnnotateHandler is a helper function that simplifies the creation of handler providers
func AnnotateHandler(fn any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			fn,
			fx.As(new(web.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
	)
}

// Module provides infrastructure dependencies
var Module = fx.Options(
	fx.Provide(
		// Core infrastructure
		echo.New,
		server.New,
		NewEventPublisher,
		// Database
		database.New,
		// Session manager
		func(logger logging.Logger) *middleware.SessionManager {
			return middleware.NewSessionManager(logger, true) // secureCookie=true for production
		},
		// Middleware manager
		func(
			core CoreParams,
			services ServiceParams,
			sessionManager *middleware.SessionManager,
		) *middleware.Manager {
			return middleware.NewManager(&middleware.ManagerConfig{
				Logger:         core.Logger,
				Security:       &core.Config.Security,
				UserService:    services.UserService,
				SessionManager: sessionManager,
				Config:         core.Config,
			})
		},
	),
)
