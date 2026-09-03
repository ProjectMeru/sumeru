package module_test

import (
	"os"
	"path/filepath"
	"testing"

	"sumeru/core/module"
	"sumeru/test/harness"
)

func TestValidateDiscoveredAddons_okMinimal(t *testing.T) {
	dir := t.TempDir()
	harness.WriteMinimalBaseAddon(t, dir)
	addon := filepath.Join(dir, "demo_x")
	if err := os.MkdirAll(filepath.Join(addon, "views"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
  "name": "demo_x",
  "version": "1.0.0",
  "depends": ["base"],
  "author": "t",
  "description": "test",
  "data": ["views/menus.xml"]
}
`
	if err := os.WriteFile(filepath.Join(addon, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	xml := `<sumeru><data></data></sumeru>`
	if err := os.WriteFile(filepath.Join(addon, "views", "menus.xml"), []byte(xml), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(addon, "init.go"), []byte(`package demo_x

import (
	_ "sumeru/base"
)
`), 0o644); err != nil {
		t.Fatal(err)
	}
	harness.WriteTempGoMod(t, dir)

	discovered, err := module.DiscoverAddonRoots([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if err := module.ValidateDiscoveredAddons(discovered); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDiscoveredAddons_modelsImportRequired(t *testing.T) {
	dir := t.TempDir()
	addon := filepath.Join(dir, "demo_y")
	if err := os.MkdirAll(filepath.Join(addon, "views"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(addon, "models"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
  "name": "demo_y",
  "version": "1.0.0",
  "depends": ["base"],
  "author": "t",
  "description": "test",
  "data": ["views/menus.xml"]
}
`
	if err := os.WriteFile(filepath.Join(addon, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	xml := `<sumeru><data></data></sumeru>`
	if err := os.WriteFile(filepath.Join(addon, "views", "menus.xml"), []byte(xml), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(addon, "init.go"), []byte("package demo_y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(addon, "models", "x.go"), []byte("package models\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	harness.WriteTempGoMod(t, dir)

	discovered, err := module.DiscoverAddonRoots([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	err = module.ValidateDiscoveredAddons(discovered)
	if err == nil {
		t.Fatal("expected error for missing models blank-import")
	}
}
