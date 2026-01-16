package framework

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/presentation/inertia"
)

func presentationModule() fx.Option {
	return fx.Module(
		"presentation",
		fx.Provide(
			inertia.NewManager,
			inertia.NewEchoHandler,
		),
	)
}
