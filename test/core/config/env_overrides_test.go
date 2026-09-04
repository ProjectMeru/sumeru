package config_test

import (
	"testing"

	"sumeru/core/server/config"
)

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("SUMERU_DB_PASSWORD", "from-env")
	t.Setenv("SUMERU_CSRF_SECRET", "csrf-from-env")
	c := &config.Config{DbPass: "ini", CSRFSecret: ""}
	config.ApplyEnvOverrides(c)
	if c.DbPass != "from-env" {
		t.Fatalf("DbPass=%q", c.DbPass)
	}
	if c.CSRFSecret != "csrf-from-env" {
		t.Fatalf("CSRFSecret=%q", c.CSRFSecret)
	}
}
