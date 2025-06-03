package bootstrap

import (
	"github.com/goformx/goforms/internal/application/handler"
	"go.uber.org/fx"
)

// HandlerProviders returns all the handler-related providers
func HandlerProviders() []fx.Option {
	return []fx.Option{
		fx.Provide(
			handler.NewAuthHandler,
			handler.NewStaticHandler,
			handler.NewVersionHandler,
			handler.NewWebHandler,
		),
	}
}
