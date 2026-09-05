package orm

import (
	"context"
	"database/sql"
	"strings"

	"sumeru/core/applog"
)

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// DBWrapper defines the interface for database operations.
// This allows the ORM to be decoupled from the standard library's sql.DB,
// facilitating better testing, logging, and risk management.
type DBWrapper interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Begin() (TxWrapper, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (TxWrapper, error)
	Close() error
	Ping() error
}

// TxWrapper defines the interface for transaction operations.
type TxWrapper interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Commit() error
	Rollback() error
}

// sqlDBWrapper implements DBWrapper using a standard *sql.DB.
type sqlDBWrapper struct {
	db *sql.DB
}

func (w *sqlDBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return w.db.Exec(query, args...)
}

func (w *sqlDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

func (w *sqlDBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return w.db.Query(query, args...)
}

func (w *sqlDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.db.QueryContext(ctx, query, args...)
}

func (w *sqlDBWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return w.db.QueryRow(query, args...)
}

func (w *sqlDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.db.QueryRowContext(ctx, query, args...)
}

func (w *sqlDBWrapper) Begin() (TxWrapper, error) {
	tx, err := w.db.Begin()
	if err != nil {
		return nil, err
	}
	return &sqlTxWrapper{tx: tx}, nil
}

func (w *sqlDBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxWrapper, error) {
	tx, err := w.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &sqlTxWrapper{tx: tx}, nil
}

func (w *sqlDBWrapper) Close() error {
	return w.db.Close()
}

func (w *sqlDBWrapper) Ping() error {
	return w.db.Ping()
}

// TableExists checks if a table exists in the current database.
// On query failure it logs and returns false only after distinguishing Scan errors
// via tableExistsErr; prefer IsInitialized for bootstrap gates.
func (w *sqlDBWrapper) TableExists(tableName string) bool {
	exists, err := w.tableExistsErr(tableName)
	if err != nil {
		applog.WarnMsg(context.Background(), "orm", "table_exists", "information_schema lookup failed", err, map[string]interface{}{"table": tableName})
		return false
	}
	return exists
}

func (w *sqlDBWrapper) tableExistsErr(tableName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE  table_schema = 'public'
		AND    table_name   = $1
	)`
	err := w.db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// IsInitialized checks if the core Sumeru tables are present.
// DB errors are not treated as "uninitialized" (avoids accidental setup re-entry).
func IsInitialized() bool {
	if DB == nil {
		return false
	}
	wrapper, ok := DB.(*sqlDBWrapper)
	if !ok {
		return false
	}
	exists, err := wrapper.tableExistsErr("sys_module")
	if err != nil {
		applog.WarnMsg(context.Background(), "orm", "is_initialized", "sys_module existence check failed; assuming initialized", err, nil)
		return true
	}
	return exists
}

// sqlTxWrapper implements TxWrapper using a standard *sql.Tx.
type sqlTxWrapper struct {
	tx *sql.Tx
}

func (w *sqlTxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return w.tx.Exec(query, args...)
}

func (w *sqlTxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.tx.ExecContext(ctx, query, args...)
}

func (w *sqlTxWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return w.tx.Query(query, args...)
}

func (w *sqlTxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.tx.QueryContext(ctx, query, args...)
}

func (w *sqlTxWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return w.tx.QueryRow(query, args...)
}

func (w *sqlTxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.tx.QueryRowContext(ctx, query, args...)
}

func (w *sqlTxWrapper) Commit() error {
	return w.tx.Commit()
}

func (w *sqlTxWrapper) Rollback() error {
	return w.tx.Rollback()
}

// NewDBWrapper creates a new DBWrapper wrapping the given *sql.DB.
func NewDBWrapper(db *sql.DB) DBWrapper {
	return &sqlDBWrapper{db: db}
}
