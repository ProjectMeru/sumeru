package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sumeru/core/server/config"
)

func minimalINI(extraLine string) string {
	base := strings.TrimSpace(`
db_host = localhost
db_port = 5432
db_user = u
db_password = p
db_name = n
http_port = 8080
addons_path = addons
`)
	if extraLine != "" {
		return base + "\n" + extraLine + "\n"
	}
	return base
}

func TestDevMode_fromINI_defaultsFalse(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sumeru.conf")
	if err := os.WriteFile(p, []byte(minimalINI("")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.LoadConfig(p); err != nil {
		t.Fatal(err)
	}
	if config.AppConfig.DevMode {
		t.Fatal("expected DevMode false when dev_mode omitted")
	}
}

func TestDevMode_fromINI_parseBoolKey(t *testing.T) {
	for _, tc := range []struct {
		line string
		want bool
	}{
		{"dev_mode = true", true},
		{"dev_mode = on", true},
		{"dev_mode = 1", true},
		{"dev_mode = yes", true},
		{"dev_mode = false", false},
		{"dev_mode = off", false},
		{"dev_mode = 0", false},
		{"dev_mode = no", false},
	} {
		t.Run(tc.line, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "sumeru.conf")
			if err := os.WriteFile(p, []byte(minimalINI(tc.line)), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := config.LoadConfig(p); err != nil {
				t.Fatal(err)
			}
			if config.AppConfig.DevMode != tc.want {
				t.Fatalf("DevMode=%v want %v", config.AppConfig.DevMode, tc.want)
			}
		})
	}
}
