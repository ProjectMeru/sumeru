package sdk

import "sumeru/core/orm"

// InitDB connects to the database (fatal on failure; matches ORM behavior).
func InitDB(dsn string) {
	orm.InitDB(dsn)
}

// SyncModels creates or updates tables for all registered models.
func SyncModels() error {
	return orm.SyncModels()
}

// DatabaseOpen reports whether the process has a database pool.
func DatabaseOpen() bool { return orm.DB != nil }
