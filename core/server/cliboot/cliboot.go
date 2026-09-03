// Package cliboot initializes config, database, and addons for standalone CLI tools.
package cliboot

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"sumeru/core/orm"
	"sumeru/core/server"
	"sumeru/core/server/config"
)

// StripLeadingArgsSeparator removes a leading "--" from os.Args (Makefile / go run convention).
func StripLeadingArgsSeparator() {
	server.StripLeadingArgsSeparator()
}

// LoadConfig loads INI config and resolves relative paths.
func LoadConfig(configPath string) error {
	if err := server.LoadConfig(configPath); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if err := server.AbsPaths(); err != nil {
		return fmt.Errorf("paths: %w", err)
	}
	return nil
}

// OpenDB opens a PostgreSQL connection using DSN from loaded config.
func OpenDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("postgres", DSN())
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	return db, nil
}

// OpenConfiguredDB loads config and returns a pinged DB handle with a timeout context.
// The caller must call cancel when done.
func OpenConfiguredDB(configPath string, timeout time.Duration) (context.Context, *sql.DB, context.CancelFunc, error) {
	if err := LoadConfig(configPath); err != nil {
		return nil, nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	db, err := OpenDB(ctx)
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}
	return ctx, db, cancel, nil
}

func initFromLoadedConfig() (context.Context, error) {
	c := config.AppConfig
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DbHost, c.DbPort, c.DbUser, c.DbPass, c.DbName, c.DbSslMode)
	server.InitDatabase(dsn)
	if !orm.IsInitialized() {
		return nil, fmt.Errorf("database %q is not initialized; complete /setup first", c.DbName)
	}
	if err := server.SyncModels(); err != nil {
		return nil, fmt.Errorf("sync models: %w", err)
	}
	if err := orm.SyncRegistrySchema(); err != nil {
		return nil, fmt.Errorf("schema sync: %w", err)
	}
	if err := server.LoadAddonPaths(config.AppConfig.AddonPaths); err != nil {
		return nil, fmt.Errorf("addons: %w", err)
	}
	ctx := orm.ContextWithBypass(context.Background(), true)
	if err := orm.EnsureDefaultGroupsAndImplied(); err != nil {
		return nil, fmt.Errorf("security groups: %w", err)
	}
	return ctx, nil
}

// Init loads INI config, connects to PostgreSQL, syncs models, and discovers addons.
// Requires an initialized database (not setup mode).
func Init(configPath string) (context.Context, error) {
	if err := LoadConfig(configPath); err != nil {
		return nil, err
	}
	return initFromLoadedConfig()
}

// InitOptionalDB loads config and addons without requiring DB (for list/depends-tree on disk only).
func InitOptionalDB(configPath string, requireDB bool) (context.Context, error) {
	if err := LoadConfig(configPath); err != nil {
		return nil, err
	}
	if err := server.LoadAddonPaths(config.AppConfig.AddonPaths); err != nil {
		return nil, fmt.Errorf("addons: %w", err)
	}
	if !requireDB {
		return context.Background(), nil
	}
	return initFromLoadedConfig()
}

// ContextWithUID returns ctx with security uid set for ORM calls.
func ContextWithUID(ctx context.Context, uid int) context.Context {
	if uid <= 0 {
		uid = 1
	}
	return orm.ContextWithUID(ctx, uid)
}

// DSN returns the primary PostgreSQL connection string from loaded config.
func DSN() string {
	c := config.AppConfig
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DbHost, c.DbPort, c.DbUser, c.DbPass, c.DbName, c.DbSslMode)
}

// ReadReplicaDSN returns read replica DSN when configured, else primary DSN.
func ReadReplicaDSN() string {
	if s := strings.TrimSpace(config.AppConfig.DbReadReplicaDSN); s != "" {
		return s
	}
	return DSN()
}
