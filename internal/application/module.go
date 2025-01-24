package application

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/http"
	"github.com/jonesrussell/goforms/internal/application/server"
	"github.com/jonesrussell/goforms/internal/domain"
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
	fx.Invoke(RegisterRoutes),
	http.Module,
	domain.Module,
)
