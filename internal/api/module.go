package api

import (
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options()
