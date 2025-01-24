package application

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/server"
)

// Module combines all application-level modules and providers
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	fx.Provide(
		NewEcho,
		server.New,
	),
	fx.Invoke(
		RegisterRoutes,
		func(srv *server.Server) {
			// Server is started via lifecycle hooks
		},
	),
)
