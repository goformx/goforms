package framework

import "go.uber.org/fx"

// Module wires application dependencies via Fx.
var Module = fx.Module(
	"framework",
	configModule(),
	infrastructureModule(),
	domainModule(),
	applicationModule(),
	middlewareModule(),
	validationModule(),
	presentationModule(),
	handlersModule(),
)
