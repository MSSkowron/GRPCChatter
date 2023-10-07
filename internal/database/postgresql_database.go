package database

import (
	"context"
	"database/sql"
	"fmt"
)

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
