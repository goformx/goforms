package web

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/web/handler"
)

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	fx.Provide(
		handler.NewPageHandler,
	),
)
