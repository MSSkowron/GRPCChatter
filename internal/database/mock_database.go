package database

import (
	"context"
	"database/sql"
)

// MockDatabase is a mock implementation of the Database interface for testing.
type MockDatabase struct {
	ExecContextFn     func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContextFn func(ctx context.Context, query string, args ...interface{}) (*sql.Row, error)
	QueryContextFn    func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	CloseFn           func() error
}

// NewMockDatabase creates a new instance of MockDatabase.
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

// ExecContext is a mock implementation of ExecContext method.
func (mdb *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if mdb.ExecContextFn != nil {
		return mdb.ExecContextFn(ctx, query, args...)
	}
	return nil, nil
}

// QueryRowContext is a mock implementation of QueryRowContext method.
func (mdb *MockDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) (*sql.Row, error) {
	if mdb.QueryRowContextFn != nil {
		return mdb.QueryRowContextFn(ctx, query, args...)
	}
	return nil, nil
}

// QueryContext is a mock implementation of QueryContext method.
func (mdb *MockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if mdb.QueryContextFn != nil {
		return mdb.QueryContextFn(ctx, query, args...)
	}
	return nil, nil
}

// Close is a mock implementation of Close method.
func (mdb *MockDatabase) Close() error {
	if mdb.CloseFn != nil {
		return mdb.CloseFn()
	}
	return nil
}
