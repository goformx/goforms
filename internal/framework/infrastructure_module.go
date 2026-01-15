package framework

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
)

func infrastructureModule() fx.Option {
	return fx.Module(
		"infrastructure",
		fx.Provide(
			infrastructure.ProvideEcho,
			provideDatabase,
			server.New,
			infrastructure.ProvideSanitizationService,
			infrastructure.NewLoggerFactory,
			infrastructure.NewLogger,
			infrastructure.NewEventPublisher,
			event.NewMemoryEventBus,
			infrastructure.ProvideAssetServer,
			infrastructure.NewAssetManager,
		),
		fx.Invoke(registerInfrastructureLifecycle),
		fx.Invoke(registerServerLifecycle),
	)
}

func provideDatabase(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger logging.Logger,
) (database.DB, error) {
	if cfg == nil {
		return nil, infrastructure.ErrMissingConfig
	}

	if logger == nil {
		return nil, infrastructure.ErrMissingLogger
	}

	db, err := database.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("Database connection established")
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info("Closing database connection")
			return db.Close()
		},
	})

	return db, nil
}

func registerInfrastructureLifecycle(lc fx.Lifecycle, logger logging.Logger, _ *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("Infrastructure module initialized")
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info("Infrastructure module shutting down")
			return nil
		},
	})
}

func registerServerLifecycle(lc fx.Lifecycle, srv *server.Server) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
