package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/config/database"

	// Import mysql driver for side effects - required for database/sql to work with MySQL
	_ "github.com/go-sql-driver/mysql"
)

func New(cfg *config.Config) (*sqlx.DB, error) {
	dsn := buildDSN(&cfg.Database)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings from config
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetme)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxLifetme) // Using same value for idle time

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// buildDSN constructs the database connection string
func buildDSN(dbConfig *database.Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName)
}
