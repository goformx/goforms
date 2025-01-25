package internal

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/http"
	"github.com/jonesrussell/goforms/internal/infrastructure"
)

// Module combines all application modules
var Module = fx.Options(
	infrastructure.Module, // DB, config, logging, auth
	domain.Module,         // Business logic services
	fx.Provide(
		http.NewHandlers,
	),
	fx.Invoke(
		registerServer, // Sets up and starts HTTP server
	),
)
