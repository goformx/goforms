package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/config"

	"time"

	_ "github.com/go-sql-driver/mysql"
)

func New(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.Database.DSN())
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
