package repositories

import (
	"database/sql"
)

// SQLiteDBWrapper wraps the SQLiteDB for use in repositories
type SQLiteDBWrapper struct {
	db *sql.DB
}

// NewSQLiteDBWrapper creates a new wrapper
func NewSQLiteDBWrapper(db *sql.DB) *SQLiteDBWrapper {
	return &SQLiteDBWrapper{db: db}
}

// GetDB returns the underlying database connection
func (w *SQLiteDBWrapper) GetDB() *sql.DB {
	return w.db
}