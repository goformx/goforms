package application

import (
	v1 "github.com/jonesrussell/goforms/internal/application/http/v1"
	"github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
)

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
