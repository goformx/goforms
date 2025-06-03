package bootstrap

import (
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// provideEcho creates a new Echo server instance
func provideEcho(logger logging.Logger) (*echo.Echo, error) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = middleware.NewValidator()
	return e, nil
}

// provideMiddlewareManager creates a new middleware manager
func provideMiddlewareManager(
	cfg *config.Config,
	userService user.Service,
	logger logging.Logger,
) *middleware.Manager {
	return middleware.New(&middleware.ManagerConfig{
		Logger:      logger,
		UserService: userService,
		Security:    &cfg.Security,
		Config:      cfg,
	})
}

// configureMiddleware sets up the middleware on the Echo instance
func configureMiddleware(e *echo.Echo, mwManager *middleware.Manager) error {
	mwManager.Setup(e)
	return nil
}

// ServerProviders returns all the server-related providers
func ServerProviders() []fx.Option {
	return []fx.Option{
		fx.Provide(
			provideEcho,
			provideMiddlewareManager,
		),
		fx.Invoke(configureMiddleware),
	}
}
