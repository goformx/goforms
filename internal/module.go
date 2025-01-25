package internal

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/http"
	"github.com/jonesrussell/goforms/internal/infrastructure"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
)

// Module combines all application modules
var Module = fx.Options(
	// Core infrastructure (DB, config, logging)
	infrastructure.Module,

	// Domain services
	domain.Module,

	// HTTP handlers
	fx.Provide(
		http.NewHandlers,
	),

	// Register routes
	fx.Invoke(func(srv *server.Server, handlers *http.Handlers) {
		handlers.Register(srv.Echo())
	}),
)
