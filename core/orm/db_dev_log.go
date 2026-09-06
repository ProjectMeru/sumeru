package orm

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"sumeru/core/applog"
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

func logSQL(ctx context.Context, op, query string, args []interface{}, err error, dur time.Duration) {
	q := strings.Join(strings.Fields(query), " ")
	if len(q) > 500 {
		q = q[:500] + "…"
	}
	ctxMap := map[string]interface{}{"op": op, "sql": q, "ms": dur.Milliseconds()}
	if len(args) > 0 {
		ctxMap["args"] = scrubSQLArgs(query, args)
	}
	status := "success"
	if err != nil {
		status = "failure"
	}
	applog.Debug(ctx, applog.Event{
		Message: "SQL", Component: "orm", Operation: "sql", Status: status, Context: ctxMap, Err: err,
	})
}

func scrubSQLArgs(query string, args []interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}
	if applog.TextContainsSecretKeyword(query) {
		return applog.RedactedPlaceholder
	}
	out := make([]interface{}, len(args))
	for i, a := range args {
		out[i] = applog.ScrubValue("", a)
	}
	return out
}

func (w *loggingDBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return w.ExecContext(context.Background(), query, args...)
}

func (w *loggingDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := w.inner.ExecContext(ctx, query, args...)
	logSQL(ctx, "exec", query, args, err, time.Since(start))
	return res, err
}

func (w *loggingDBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return w.QueryContext(context.Background(), query, args...)
}

func (w *loggingDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.inner.QueryContext(ctx, query, args...)
	logSQL(ctx, "query", query, args, err, time.Since(start))
	return rows, err
}

func (w *loggingDBWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return w.QueryRowContext(context.Background(), query, args...)
}

func (w *loggingDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := w.inner.QueryRowContext(ctx, query, args...)
	logSQL(ctx, "query_row", query, args, nil, time.Since(start))
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
	return &loggingTxWrapper{inner: tx}, nil
}

func (w *loggingDBWrapper) Close() error { return w.inner.Close() }
func (w *loggingDBWrapper) Ping() error  { return w.inner.Ping() }

type loggingTxWrapper struct {
	inner TxWrapper
}

func (w *loggingTxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return w.ExecContext(context.Background(), query, args...)
}

func (w *loggingTxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := w.inner.ExecContext(ctx, query, args...)
	logSQL(ctx, "tx_exec", query, args, err, time.Since(start))
	return res, err
}

func (w *loggingTxWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return w.QueryContext(context.Background(), query, args...)
}

func (w *loggingTxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.inner.QueryContext(ctx, query, args...)
	logSQL(ctx, "tx_query", query, args, err, time.Since(start))
	return rows, err
}

func (w *loggingTxWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return w.QueryRowContext(context.Background(), query, args...)
}

func (w *loggingTxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := w.inner.QueryRowContext(ctx, query, args...)
	logSQL(ctx, "tx_query_row", query, args, nil, time.Since(start))
	return row
}

func (w *loggingTxWrapper) Commit() error   { return w.inner.Commit() }
func (w *loggingTxWrapper) Rollback() error { return w.inner.Rollback() }
