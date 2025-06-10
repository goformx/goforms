package application

import (
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/services/auth"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
)

// ServiceParams contains dependencies for application services
type ServiceParams struct {
	fx.In

	UserService user.Service
	FormService form.Service
	Logger      logging.Logger
}

// HandlerParams contains dependencies for handlers
type HandlerParams struct {
	fx.In

	UserService       user.Service
	FormService       form.Service
	AuthService       auth.Service
	SessionManager    *middleware.SessionManager
	Renderer          *view.Renderer
	MiddlewareManager *middleware.Manager
	Config            *config.Config
	Logger            logging.Logger
}

// NewAuthService creates a new auth service with proper error handling
func NewAuthService(p ServiceParams) (auth.Service, error) {
	if p.UserService == nil {
		return nil, errors.New("user service is required for auth service")
	}
	if p.Logger == nil {
		return nil, errors.New("logger is required for auth service")
	}
	return auth.NewService(p.UserService, p.Logger), nil
}

// NewHandlerDeps creates a new HandlerDeps instance with proper error handling
func NewHandlerDeps(p HandlerParams) (*web.HandlerDeps, error) {
	deps := &web.HandlerDeps{
		UserService:       p.UserService,
		FormService:       p.FormService,
		AuthService:       p.AuthService,
		SessionManager:    p.SessionManager,
		Renderer:          p.Renderer,
		MiddlewareManager: p.MiddlewareManager,
		Config:            p.Config,
		Logger:            p.Logger,
	}

	// Validate all required dependencies
	if err := deps.Validate(
		"UserService",
		"FormService",
		"AuthService",
		"SessionManager",
		"Renderer",
		"MiddlewareManager",
		"Config",
		"Logger",
	); err != nil {
		return nil, err
	}

	return deps, nil
}

// NewAuthHandler creates a new auth handler with proper error handling
func NewAuthHandler(deps *web.HandlerDeps) (*web.AuthHandler, error) {
	return web.NewAuthHandler(*deps)
}

// NewWebHandler creates a new web handler with proper error handling
func NewWebHandler(deps *web.HandlerDeps) (*web.WebHandler, error) {
	return web.NewWebHandler(*deps)
}

// NewFormHandler creates a new form handler with proper error handling
func NewFormHandler(deps *web.HandlerDeps, formService form.Service) (*web.FormHandler, error) {
	handler := web.NewFormHandler(*deps, formService)
	return handler, nil
}

// NewDemoHandler creates a new demo handler with proper error handling
func NewDemoHandler(deps *web.HandlerDeps) (*web.DemoHandler, error) {
	return web.NewDemoHandler(*deps)
}

// Module provides all application layer dependencies
var Module = fx.Options(
	// Services
	fx.Provide(
		fx.Annotate(
			NewAuthService,
			fx.As(new(auth.Service)),
		),
	),
	// Handler dependencies
	fx.Provide(
		NewHandlerDeps,
	),
	// Handlers
	fx.Provide(
		fx.Annotate(
			NewAuthHandler,
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(web.Handler)),
		),
		fx.Annotate(
			NewWebHandler,
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(web.Handler)),
		),
		fx.Annotate(
			NewFormHandler,
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(web.Handler)),
		),
		fx.Annotate(
			NewDemoHandler,
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(web.Handler)),
		),
	),
)
