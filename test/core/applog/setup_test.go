package applog_test

import (
	"path/filepath"
	"testing"

	"sumeru/core/applog"
	"sumeru/core/server/config"
)

func TestSetupFromConfig_rollingRequiresFile(t *testing.T) {
	cfg := &config.Config{
		LogEnabled: true,
		LogStdout:  true,
		LogRolling: true,
		LogFile:    "",
	}
	if err := applog.SetupFromConfig(cfg); err == nil {
		t.Fatal("expected error when log_rolling without log_file")
	}
}

func TestSetupFromConfig_stdoutOnly(t *testing.T) {
	cfg := &config.Config{
		LogEnabled: true,
		LogStdout:  true,
		LogFile:    "",
	}
	if err := applog.SetupFromConfig(cfg); err != nil {
		t.Fatal(err)
	}
	defer applog.Sync()
}

func TestSetupFromConfig_fileAppend(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		LogEnabled: true,
		LogStdout:  false,
		LogRolling: false,
		LogFile:    filepath.Join(dir, "app.log"),
	}
	if err := applog.SetupFromConfig(cfg); err != nil {
		t.Fatal(err)
	}
	defer applog.Sync()
}
