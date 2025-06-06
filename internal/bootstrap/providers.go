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
	logger, err := factory.CreateLogger()
	if err != nil {
		return nil, err
	}
	logger.Info("Logger initialized successfully")
	return logger, nil
}

// provideZapLogger creates a new zap logger instance
func provideZapLogger(logger logging.Logger) *zap.Logger {
	logger.Info("Initializing Zap logger")
	if zapLogger, ok := logger.(*logging.ZapLogger); ok {
		logger.Info("Using existing Zap logger instance")
		return zapLogger.GetZapLogger()
	}
	logger.Info("Creating new development Zap logger")
	devLogger, _ := zap.NewDevelopment()
	return devLogger
}

// provideFxLogger creates a new fx logger instance
func provideFxLogger(log *zap.Logger) fxevent.Logger {
	log.Info("Initializing Fx logger")
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
