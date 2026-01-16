package framework

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

func middlewareModule() fx.Option {
	return fx.Module(
		"middleware",
		fx.Provide(
			constants.NewPathManager,
			auth.NewMiddleware,
			provideAccessManager,
			provideSessionManager,
			provideMiddlewareConfig,
			provideRegistry,
			provideOrchestrator,
			middleware.NewEchoOrchestratorAdapter,
			middleware.NewMigrationAdapter,
			provideLegacyManager,
		),
		fx.Invoke(registerMiddlewareSystem),
		fx.Invoke(registerSessionLifecycle),
	)
}

func provideAccessManager(_ logging.Logger, pathManager *constants.PathManager) *access.Manager {
	accessConfig := &access.Config{
		DefaultAccess: access.Authenticated,
		PublicPaths:   pathManager.PublicPaths,
		AdminPaths:    pathManager.AdminPaths,
	}
	rules := middleware.GenerateAccessRules(pathManager)

	return access.NewManager(accessConfig, rules)
}

func provideSessionManager(
	logger logging.Logger,
	cfg *config.Config,
	accessManager *access.Manager,
	pathManager *constants.PathManager,
) *session.Manager {
	sessionConfig := &session.Config{
		SessionConfig: &cfg.Session,
		Config:        cfg,
		PublicPaths:   pathManager.PublicPaths,
		StaticPaths:   pathManager.StaticPaths,
	}

	return session.NewManager(logger, sessionConfig, accessManager)
}

func provideMiddlewareConfig(cfg *config.Config, logger logging.Logger) middleware.MiddlewareConfig {
	return middleware.NewMiddlewareConfig(cfg, logger)
}

func provideRegistry(logger logging.Logger, cfg middleware.MiddlewareConfig) core.Registry {
	return middleware.NewRegistry(logger, cfg)
}

func provideOrchestrator(
	registry core.Registry,
	cfg middleware.MiddlewareConfig,
	logger logging.Logger,
) core.Orchestrator {
	return middleware.NewOrchestrator(registry, cfg, logger)
}

func provideLegacyManager(
	logger logging.Logger,
	cfg *config.Config,
	userService user.Service,
	formService form.Service,
	sessionManager *session.Manager,
	accessManager *access.Manager,
	sanitizer sanitization.ServiceInterface,
) *middleware.Manager {
	return middleware.NewManager(&middleware.ManagerConfig{
		Logger:         logger,
		Config:         cfg,
		UserService:    userService,
		FormService:    formService,
		SessionManager: sessionManager,
		AccessManager:  accessManager,
		Sanitizer:      sanitizer,
	})
}

func registerMiddlewareSystem(
	lc fx.Lifecycle,
	registry core.Registry,
	orchestrator core.Orchestrator,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := middleware.RegisterAllMiddleware(registry, logger); err != nil {
				return err
			}

			if err := orchestrator.ValidateConfiguration(); err != nil {
				return fmt.Errorf("failed to validate orchestrator configuration: %w", err)
			}

			logger.Info("middleware system initialized successfully")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("middleware system shutting down")
			return nil
		},
	})
}

func registerSessionLifecycle(lc fx.Lifecycle, manager *session.Manager) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return manager.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return manager.Stop(ctx)
		},
	})
}
