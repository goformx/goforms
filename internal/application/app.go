package application

import (
	"github.com/labstack/echo/v4"

	v1 "github.com/jonesrussell/goforms/internal/application/http/v1"
	"github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
	"go.uber.org/fx"
)

// App represents the application
type App struct {
	echo   *echo.Echo
	logger logging.Logger
}

// Start creates and starts a new application
func Start() *fx.App {
	return fx.New(
		fx.Provide(
			config.New,
			logging.New,
			server.New,
			database.New,
			v1.Module,
			middleware.New,
		),
		fx.Invoke(registerHandlers),
	)
}

func registerHandlers(
	srv *server.Server,
	contactAPI *v1.ContactAPI,
	subscriptionAPI *v1.SubscriptionAPI,
	handler *v1.Handler,
	mw *middleware.Manager,
	logger logging.Logger,
) {
	e := srv.Echo()

	// Setup middleware
	mw.Setup(e)
	e.Use(middleware.LoggingMiddleware(logger))

	// Register API routes
	contactAPI.Register(e)
	subscriptionAPI.Register(e)

	// Register web routes
	handler.Register(e)
}
