package database

import (
	"context"
	"fmt"

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

// PostgresDatabase implements the Database interface for PostgreSQL.
type PostgresDatabase struct {
	db *sql.DB
}

// NewPostgresDatabase creates a new PostgresDatabase instance with the connection string and context.
// It establishes a connection to the PostgreSQL database and verifies its availability.
func NewPostgresDatabase(ctx context.Context, connectionString string) (*PostgresDatabase, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to check database connection: %w", err)
	}

	return &PostgresDatabase{
		db: db,
	}, nil
}

func (pdb *PostgresDatabase) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return pdb.db.ExecContext(ctx, query, args...)
}

func (pdb *PostgresDatabase) QueryRowContext(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	return pdb.db.QueryRowContext(ctx, query, args...), nil
}

func (pdb *PostgresDatabase) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return pdb.db.QueryContext(ctx, query, args...)
}

func (pdb *PostgresDatabase) Close() error {
	if err := pdb.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}
