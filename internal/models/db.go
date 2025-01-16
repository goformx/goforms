package models

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DB is a database interface that wraps the standard database/sql methods
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) Row
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Result is a result interface that wraps the standard database/sql Result interface
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Row is a row interface that wraps the standard database/sql Row interface
type Row interface {
	Scan(dest ...interface{}) error
}

// Compile-time check to ensure we're using methods that exist in sqlx.DB
var _ interface {
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} = (*sqlx.DB)(nil)
