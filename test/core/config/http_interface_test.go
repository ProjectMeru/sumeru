package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"sumeru/core/server/config"
)

func TestLoadHTTPInterface(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sumeru.conf")
	content := `db_host = localhost
db_port = 5432
db_user = u
db_password = p
db_name = n
http_port = 8080
http_interface = 127.0.0.1
addons_path = addons
`
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.LoadConfig(p); err != nil {
		t.Fatal(err)
	}
	if config.AppConfig.HttpInterface != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %q", config.AppConfig.HttpInterface)
	}
}
