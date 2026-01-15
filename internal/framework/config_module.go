package framework

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/infrastructure/config"
)

func configModule() fx.Option {
	return fx.Module(
		"config",
		fx.Provide(
			provideConfig,
			config.NewAppConfig,
			config.NewDatabaseConfig,
			config.NewSecurityConfig,
			config.NewEmailConfig,
			config.NewStorageConfig,
			config.NewCacheConfig,
			config.NewLoggingConfig,
			config.NewSessionConfig,
			config.NewAuthConfig,
			config.NewFormConfig,
			config.NewAPIConfig,
			config.NewWebConfig,
			config.NewUserConfig,
		),
	)
}

func provideConfig() (*config.Config, error) {
	viperConfig := config.NewViperConfig()

	return viperConfig.Load()
}
