package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

func New(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.Database.DSN())
	if err != nil {
		return nil, err
	}

	// Set connection pool settings from config
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxLifetime) // Using same value for idle time

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
