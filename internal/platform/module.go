package platform

import (
	"os"
	"strconv"

	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/platform/database"
)

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	fx.Provide(
		NewDatabaseConfig,
		database.New,
		database.NewContactStore,
		logger.GetLogger,
	),
)

func NewDatabaseConfig() (database.Config, error) {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		port = 3306 // default MySQL port
	}

	return database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	}, nil
}
