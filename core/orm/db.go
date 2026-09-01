package orm

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"sumeru/core/applog"
)

// DBPoolSettings configures the PostgreSQL connection pool.
type DBPoolSettings struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

var DB DBWrapper
var readDB DBWrapper
var readReplicaReady bool

func InitDB(connStr string) {
	InitDBWithPool(connStr, DBPoolSettings{})
}

// InitDBWithPool opens the primary pool and optionally configures limits.
func InitDBWithPool(connStr string, pool DBPoolSettings) {
	rawDB, err := sql.Open("postgres", connStr)
	if err != nil {
		applog.Fatal(context.Background(), "db_open", "err", err)
	}
	applyPoolSettings(rawDB, pool)
	DB = wrapDevLogging(NewDBWrapper(rawDB))
	if err := DB.Ping(); err != nil {
		applog.Fatal(context.Background(), "db_ping", "err", err)
	}
	applog.InfoMsg(context.Background(), "orm", "connect", "Successfully connected to the database", nil)
}

// InitReadReplica opens an optional read-only replica pool for search-heavy paths.
func InitReadReplica(connStr string, pool DBPoolSettings) {
	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		readDB = nil
		readReplicaReady = false
		return
	}
	rawDB, err := sql.Open("postgres", connStr)
	if err != nil {
		applog.Fatal(context.Background(), "db_read_open", "err", err)
	}
	applyPoolSettings(rawDB, pool)
	readDB = wrapDevLogging(NewDBWrapper(rawDB))
	if err := readDB.Ping(); err != nil {
		applog.Fatal(context.Background(), "db_read_ping", "err", err)
	}
	readReplicaReady = true
	applog.InfoMsg(context.Background(), "orm", "connect", "Read replica connected", nil)
}

func applyPoolSettings(db *sql.DB, pool DBPoolSettings) {
	if pool.MaxOpenConns > 0 {
		db.SetMaxOpenConns(pool.MaxOpenConns)
	}
	if pool.MaxIdleConns > 0 {
		db.SetMaxIdleConns(pool.MaxIdleConns)
	}
	if pool.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(pool.ConnMaxLifetime)
	}
}

type readReplicaKey struct{}

// ContextWithReadReplica marks ctx to prefer the read replica for Search* queries.
func ContextWithReadReplica(ctx context.Context, use bool) context.Context {
	return context.WithValue(ctx, readReplicaKey{}, use)
}

func useReadReplica(ctx context.Context) bool {
	v, _ := ctx.Value(readReplicaKey{}).(bool)
	return v && readReplicaReady
}

// QueryDB returns read replica when ctx requests it and replica is configured.
func QueryDB(ctx context.Context) DBWrapper {
	if useReadReplica(ctx) {
		return readDB
	}
	return DB
}
