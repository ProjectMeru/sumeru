package server

import (
	"sumeru/core/module"
	"sumeru/core/sdk"
	"sumeru/core/server/config"
)

// Config is the on-disk configuration shape (stable alias via sdk).
type Config = sdk.Config

// LoadConfig loads INI configuration from path.
func LoadConfig(path string) error {
	return config.LoadConfig(path)
}

// AbsPaths resolves relative paths in AppConfig.
func AbsPaths() error {
	return config.AbsPaths()
}

// InitDB opens the database pool.
func InitDB(databaseConnectionString string) {
	sdk.InitDB(databaseConnectionString)
}

// SyncModels syncs registered models to the database.
func SyncModels() error {
	return sdk.SyncModels()
}

// LoadAddonPaths discovers and loads addons from the given roots.
func LoadAddonPaths(paths []string) error {
	return module.LoadAddonPaths(paths)
}

// RunModuleCLI runs -i / -u module lists.
func RunModuleCLI(installCSV, updateCSV string) error {
	return module.RunModuleCLI(installCSV, updateCSV)
}
