package database

import (
	"context"
	"fmt"

	"database/sql"

	_ "github.com/lib/pq"
)

type Database interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) (*sql.Row, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Close() error
}

// PostgresDatabase implements the Database interface for PostgreSQL.
type PostgresDatabase struct {
	db *sql.DB
}

// NewPostgresDatabase creates a new PostgresDatabase instance.
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

// ExecContext executes a query that doesn't return rows.
func (pdb *PostgresDatabase) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return pdb.db.ExecContext(ctx, query, args...)
}

// QueryRowContext retrieves a single row.
func (pdb *PostgresDatabase) QueryRowContext(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	return pdb.db.QueryRowContext(ctx, query, args...), nil
}

// QueryContext executes a query that returns multiple rows.
func (pdb *PostgresDatabase) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return pdb.db.QueryContext(ctx, query, args...)
}

// Close closes the database connection.
func (pdb *PostgresDatabase) Close() error {
	if err := pdb.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}
