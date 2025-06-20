package validation

import (
	infra_validation "github.com/goformx/goforms/internal/infrastructure/validation"
	"go.uber.org/fx"
)

// Module provides validation dependencies
var Module = fx.Options(
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
