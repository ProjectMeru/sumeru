package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"sumeru/core/server/config"
)

func TestLoadConfigExtendedKeys(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sumeru.conf")
	ini := minimalINI(`
dev_features = sql,access
log_timezone = UTC
log_enabled = false
setup_token = secret
rate_limit_rpm = 120
smtp_host = smtp.test
smtp_port = 465
smtp_user = user
smtp_password = pass
smtp_from = noreply@test
db_max_open_conns = 10
db_max_idle_conns = 5
db_conn_max_lifetime_minutes = 30
`)
	if err := os.WriteFile(p, []byte(ini), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.LoadConfig(p); err != nil {
		t.Fatal(err)
	}
	c := config.AppConfig
	if c.DevFeatures != "sql,access" || c.LogTimezone != "UTC" || c.LogEnabled {
		t.Fatalf("features/timezone/log: %+v", c)
	}
	if c.SetupToken != "secret" || c.RateLimitRPM != 120 {
		t.Fatalf("setup/rate: %+v", c)
	}
	if c.SMTPHost != "smtp.test" || c.SMTPPort != 465 || c.SMTPFrom != "noreply@test" {
		t.Fatalf("smtp: %+v", c)
	}
	if c.DbMaxOpenConns != 10 || c.DbMaxIdleConns != 5 || c.DbConnMaxLifetimeMin != 30 {
		t.Fatalf("pool: %+v", c)
	}
}

func TestAbsPaths(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sumeru.conf")
	if err := os.WriteFile(p, []byte(minimalINI("addons_path = addons,extra")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.LoadConfig(p); err != nil {
		t.Fatal(err)
	}
	if err := config.AbsPaths(); err != nil {
		t.Fatal(err)
	}
	if len(config.AppConfig.AddonPaths) != 2 {
		t.Fatalf("addon paths: %v", config.AppConfig.AddonPaths)
	}
}

func TestParseDevFeatures(t *testing.T) {
	flags := config.ParseDevFeatures("sql, access ,XML")
	if !flags["sql"] || !flags["access"] || !flags["xml"] {
		t.Fatalf("flags=%v", flags)
	}
}
