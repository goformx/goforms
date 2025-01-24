package user

import (
	"go.uber.org/fx"
)

// Module provides user domain dependencies
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewService,
			fx.ParamTags(``, ``, `name:"jwt_secret"`),
		),
	),
)
