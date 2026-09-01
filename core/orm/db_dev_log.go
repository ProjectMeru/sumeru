package orm

import (
	"context"
	"database/sql"
	"strings"
	"time"

	applog "sumeru/core/applog"
)

type loggingDBWrapper struct {
	inner DBWrapper
}

func wrapDevLogging(db DBWrapper) DBWrapper {
	if !DevFeatureEnabled("sql") {
		return db
	}
	return &loggingDBWrapper{inner: db}
}

type sqlLogEntry struct {
	Op    string
	Query string
	Args  []interface{}
	Err   error
	Dur   time.Duration
}

func logSQL(ctx context.Context, in sqlLogEntry) {
	if !DevFeatureEnabled("sql") {
		return
	}
	q := strings.Join(strings.Fields(in.Query), " ")
	if len(q) > 500 {
		q = q[:500] + "…"
	}
	fields := []interface{}{"op", in.Op, "sql", q, "ms", in.Dur.Milliseconds()}
	if len(in.Args) > 0 {
		fields = append(fields, "args", in.Args)
	}
	if in.Err != nil {
		fields = append(fields, "err", in.Err.Error())
	}
	applog.L(ctx).Debug("dev_sql", fields...)
}

func (w *loggingDBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := w.inner.Exec(query, args...)
	logSQL(context.Background(), sqlLogEntry{Op: "exec", Query: query, Args: args, Err: err, Dur: time.Since(start)})
	return res, err
}

func (w *loggingDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := w.inner.ExecContext(ctx, query, args...)
	logSQL(ctx, sqlLogEntry{Op: "exec", Query: query, Args: args, Err: err, Dur: time.Since(start)})
	return res, err
}

func (w *loggingDBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.inner.Query(query, args...)
	logSQL(context.Background(), sqlLogEntry{Op: "query", Query: query, Args: args, Err: err, Dur: time.Since(start)})
	return rows, err
}

func (w *loggingDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.inner.QueryContext(ctx, query, args...)
	logSQL(ctx, sqlLogEntry{Op: "query", Query: query, Args: args, Err: err, Dur: time.Since(start)})
	return rows, err
}

func (w *loggingDBWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := w.inner.QueryRow(query, args...)
	logSQL(context.Background(), sqlLogEntry{Op: "query_row", Query: query, Args: args, Dur: time.Since(start)})
	return row
}

func (w *loggingDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := w.inner.QueryRowContext(ctx, query, args...)
	logSQL(ctx, sqlLogEntry{Op: "query_row", Query: query, Args: args, Dur: time.Since(start)})
	return row
}

func (w *loggingDBWrapper) Begin() (TxWrapper, error) {
	tx, err := w.inner.Begin()
	if err != nil {
		return nil, err
	}
	return &loggingTxWrapper{inner: tx}, nil
}

func (w *loggingDBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxWrapper, error) {
	tx, err := w.inner.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &loggingTxWrapper{inner: tx, ctx: ctx}, nil
}

func (w *loggingDBWrapper) Close() error  { return w.inner.Close() }
func (w *loggingDBWrapper) Ping() error   { return w.inner.Ping() }

type loggingTxWrapper struct {
	inner TxWrapper
	ctx   context.Context
}

func (w *loggingTxWrapper) txCtx() context.Context {
	if w.ctx != nil {
		return w.ctx
	}
	return context.Background()
}

func (w *loggingTxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := w.inner.Exec(query, args...)
	logSQL(w.txCtx(), sqlLogEntry{Op: "tx_exec", Query: query, Args: args, Err: err, Dur: time.Since(start)})
	return res, err
}

func (w *loggingTxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := w.inner.ExecContext(ctx, query, args...)
	logSQL(ctx, sqlLogEntry{Op: "tx_exec", Query: query, Args: args, Err: err, Dur: time.Since(start)})
	return res, err
}

func (w *loggingTxWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.inner.Query(query, args...)
	logSQL(w.txCtx(), sqlLogEntry{Op: "tx_query", Query: query, Args: args, Err: err, Dur: time.Since(start)})
	return rows, err
}

func (w *loggingTxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.inner.QueryContext(ctx, query, args...)
	logSQL(ctx, sqlLogEntry{Op: "tx_query", Query: query, Args: args, Err: err, Dur: time.Since(start)})
	return rows, err
}

func (w *loggingTxWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := w.inner.QueryRow(query, args...)
	logSQL(w.txCtx(), sqlLogEntry{Op: "tx_query_row", Query: query, Args: args, Dur: time.Since(start)})
	return row
}

func (w *loggingTxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := w.inner.QueryRowContext(ctx, query, args...)
	logSQL(ctx, sqlLogEntry{Op: "tx_query_row", Query: query, Args: args, Dur: time.Since(start)})
	return row
}

func (w *loggingTxWrapper) Commit() error   { return w.inner.Commit() }
func (w *loggingTxWrapper) Rollback() error { return w.inner.Rollback() }
