package internal

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/infrastructure"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

// Module combines all application modules
var Module = fx.Options(
	// Core infrastructure (DB, config, logging)
	infrastructure.Module,

	// Domain services
	domain.Module,

	// View renderer
	view.Module,
)
