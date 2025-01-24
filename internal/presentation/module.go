package web

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/presentation/view"
)

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	view.Module,
)
