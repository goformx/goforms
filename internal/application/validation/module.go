package validation

import (
	"go.uber.org/fx"

	infra_validation "github.com/goformx/goforms/internal/infrastructure/validation"
)

// Module provides validation dependencies
var Module = fx.Module("validation",
	fx.Provide(
		// Core validator
		infra_validation.New,

		// Schema generator
		NewSchemaGenerator,

		// Form validator
		fx.Annotate(
			NewFormValidator,
		),
	),
)
