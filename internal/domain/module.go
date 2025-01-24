package domain

import (
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
)

// Module combines all domain modules
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	contact.Module,
	subscription.Module,
	user.Module,
)
