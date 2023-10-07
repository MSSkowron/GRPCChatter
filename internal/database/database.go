package database

import (
	"context"

	"database/sql"

	_ "github.com/lib/pq"
)

// Database is an interface that defines the methods required for interacting with a database.
type Database interface {
	// ExecContext executes a query that doesn't return rows.
	// It should be used for INSERT, UPDATE, DELETE, and other non-query SQL statements.
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	// QueryRowContext retrieves a single row from the database.
	// It should be used for querying a single row result.
	QueryRowContext(ctx context.Context, query string, args ...any) (*sql.Row, error)

	// QueryContext executes a query that returns multiple rows.
	// It should be used for querying multiple rows of data.
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// Close closes the database connection.
	// It should be called when you're done using the database to release resources.
	Close() error
}
