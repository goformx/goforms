package models

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DB is a database interface that wraps the standard database/sql methods
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Compile-time check to ensure we're using methods that exist in sqlx.DB
var _ DB = (*sqlx.DB)(nil)
