// Package harness provides shared helpers for tests under sumeru/test/.
package harness

import (
	"os"
	"path/filepath"
	"testing"

	"sumeru/core/modelreg"
	"sumeru/core/module"
	"sumeru/core/server/config"
)

// RepoRoot returns the absolute path to the sumeru module root (directory containing go.mod).
func RepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := module.FindRepoRoot(wd)
	if err != nil {
		t.Fatalf("FindRepoRoot: %v", err)
	}
	return root
}

// ActivateModules registers queued models in manifest dependency order.
// Blank-import addon model packages before calling when addon models are required.
func ActivateModules(t *testing.T, order ...string) {
	t.Helper()
	if err := modelreg.ActivateAll(order); err != nil {
		t.Fatalf("modelreg.ActivateAll(%v): %v", order, err)
	}
}

// DiscoverFromExampleConfig loads sumeru.conf.example and discovers addons on addons_path.
func DiscoverFromExampleConfig(t *testing.T) map[string]*module.Addon {
	t.Helper()
	root := RepoRoot(t)
	cfgPath := filepath.Join(root, "sumeru.conf.example")
	if err := config.LoadConfig(cfgPath); err != nil {
		t.Fatalf("LoadConfig(%s): %v", cfgPath, err)
	}
	if err := config.AbsPaths(); err != nil {
		t.Fatalf("AbsPaths: %v", err)
	}
	discovered, err := module.DiscoverAddonRoots(config.AppConfig.AddonPaths)
	if err != nil {
		t.Fatalf("DiscoverAddonRoots: %v", err)
	}
	return discovered
}

// TempAddon writes a minimal addon tree under dir and returns its path.
type TempAddon struct {
	Name string
	Path string
	Dir  string
}

// WriteMinimalBaseAddon creates a stub base addon under dir for convention tests.
func WriteMinimalBaseAddon(t *testing.T, dir string) TempAddon {
	t.Helper()
	return writeMinimalAddon(t, dir, "base", nil, "package base\n")
}

// WriteMinimalAddon creates a stub addon with optional depends and init.go body.
func WriteMinimalAddon(t *testing.T, dir, name string, depends []string, initBody string) TempAddon {
	t.Helper()
	return writeMinimalAddon(t, dir, name, depends, initBody)
}

func writeMinimalAddon(t *testing.T, dir, name string, depends []string, initBody string) TempAddon {
	t.Helper()
	addonPath := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Join(addonPath, "views"), 0o755); err != nil {
		t.Fatal(err)
	}
	depJSON := "[]"
	if len(depends) > 0 {
		depJSON = `["` + depends[0] + `"`
		for _, d := range depends[1:] {
			depJSON += `,"` + d + `"`
		}
		depJSON += "]"
	}
	manifest := `{
  "name": "` + name + `",
  "version": "1.0.0",
  "depends": ` + depJSON + `,
  "author": "t",
  "description": "test",
  "data": ["views/menus.xml"]
}
`
	if err := os.WriteFile(filepath.Join(addonPath, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	xml := `<sumeru><data></data></sumeru>`
	if err := os.WriteFile(filepath.Join(addonPath, "views", "menus.xml"), []byte(xml), 0o644); err != nil {
		t.Fatal(err)
	}
	if initBody == "" {
		initBody = "package " + name + "\n"
	}
	if err := os.WriteFile(filepath.Join(addonPath, "init.go"), []byte(initBody), 0o644); err != nil {
		t.Fatal(err)
	}
	return TempAddon{Name: name, Path: addonPath, Dir: dir}
}

// WriteTempGoMod writes a minimal go.mod in dir for temp addon discovery tests.
func WriteTempGoMod(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module sumeru\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
