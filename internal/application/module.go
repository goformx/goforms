package application

import (
	"go.uber.org/fx"

	v1 "github.com/jonesrussell/goforms/internal/application/http/v1"
	"github.com/jonesrussell/goforms/internal/application/server"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/labstack/echo/v4"
)

// Module combines all application-level modules and providers
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	fx.Provide(
		NewEcho,
		server.New,
		// Provide contact service implementation as the interface
		fx.Annotate(
			func(impl *contact.ServiceImpl) contact.Service { return impl },
			fx.As(new(contact.Service)),
		),
	),
	fx.Invoke(
		func(e *echo.Echo, h *v1.Handler) {
			RegisterRoutes(e, h)
		},
		func(srv *server.Server) {
			// Server is started via lifecycle hooks
		},
	),
)
