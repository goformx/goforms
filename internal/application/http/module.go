package http

import (
	"go.uber.org/fx"

	v1 "github.com/jonesrussell/goforms/internal/application/http/v1"
)

// Module combines all HTTP-related modules
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	v1.Module,
)
