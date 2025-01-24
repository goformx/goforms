package application

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/http"
	v1 "github.com/jonesrussell/goforms/internal/application/http/v1"
	"github.com/jonesrussell/goforms/internal/application/server"
	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/labstack/echo/v4"
)

// Module combines all application-level modules and providers
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	fx.Provide(
		NewEcho,
		NewServerConfig,
		server.New,
	),
	fx.Invoke(func(
		e *echo.Echo,
		contactAPI *v1.ContactAPI,
		subscriptionAPI *v1.SubscriptionAPI,
		webHandler *v1.WebHandler,
	) {
		RegisterRoutes(e, contactAPI, subscriptionAPI, webHandler)
	}),
	http.Module,
	domain.Module,
)
