package bootstrap

import (
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// provideLogger creates a new logger instance
func provideLogger() (logging.Logger, error) {
	factory := logging.NewFactory()
	return factory.CreateLogger()
}

// provideZapLogger creates a new zap logger instance
func provideZapLogger(logger logging.Logger) *zap.Logger {
	if zapLogger, ok := logger.(*logging.ZapLogger); ok {
		return zapLogger.GetZapLogger()
	}
	devLogger, _ := zap.NewDevelopment()
	return devLogger
}

// provideFxLogger creates a new fx logger instance
func provideFxLogger(log *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: log}
}

// Providers returns all the providers needed for the application
func Providers() []fx.Option {
	return []fx.Option{
		fx.Provide(
			provideLogger,
			provideZapLogger,
		),
		fx.WithLogger(provideFxLogger),
	}
}
