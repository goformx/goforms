package framework

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/validation"
	infra_validation "github.com/goformx/goforms/internal/infrastructure/validation"
)

func validationModule() fx.Option {
	return fx.Module(
		"validation",
		fx.Provide(
			infra_validation.New,
			validation.NewSchemaGenerator,
			validation.NewFormValidator,
		),
	)
}
