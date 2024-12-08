package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/config"
	_ "github.com/lib/pq"
)

func New(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
