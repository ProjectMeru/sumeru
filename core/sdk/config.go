package sdk

import (
	"context"

	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// Config is the on-disk configuration shape (INI [options]).
type Config = config.Config

// Cfg returns the live process configuration (same memory as the INI-backed singleton).
func Cfg() *Config {
	return &config.AppConfig
}

// GetConfigParam returns a stored parameter, or defaultVal when it is missing.
func GetConfigParam(ctx context.Context, key, defaultVal string) string {
	return orm.GetConfigParam(ctx, key, defaultVal)
}
