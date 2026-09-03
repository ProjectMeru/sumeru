package module_test

import (
	"path/filepath"
	"testing"

	"sumeru/core/module"
	"sumeru/test/harness"
)

func TestFindRepoRootAndModulePath(t *testing.T) {
	root := harness.RepoRoot(t)
	if got, err := module.FindRepoRoot(root); err != nil || got != root {
		t.Fatalf("FindRepoRoot: %q err=%v", got, err)
	}
	modPath, err := module.ReadModulePath(root)
	if err != nil || modPath != "sumeru" {
		t.Fatalf("ReadModulePath: %q err=%v", modPath, err)
	}
}

func TestValidateDiscoveredDepends(t *testing.T) {
	dir := t.TempDir()
	base := harness.WriteMinimalBaseAddon(t, dir)
	child := harness.WriteMinimalAddon(t, dir, "child", []string{"base"}, "package child\n")
	discovered := map[string]*module.Addon{
		"base":  {Manifest: module.Manifest{Name: "base", Depends: []string{}}, Path: base.Path},
		"child": {Manifest: module.Manifest{Name: "child", Depends: []string{"base"}}, Path: child.Path},
	}
	if err := module.ValidateDiscoveredDepends(discovered); err != nil {
		t.Fatal(err)
	}
	discovered["child"].Manifest.Depends = []string{"missing"}
	if err := module.ValidateDiscoveredDepends(discovered); err == nil {
		t.Fatal("expected missing dep error")
	}
}

func TestModuleNamesTopo(t *testing.T) {
	discovered := map[string]*module.Addon{
		"base": {Manifest: module.Manifest{Name: "base"}},
		"sale": {Manifest: module.Manifest{Name: "sale", Depends: []string{"base"}}},
	}
	names, err := module.ModuleNamesTopo(discovered)
	if err != nil || len(names) != 2 || names[0] != "base" {
		t.Fatalf("names=%v err=%v", names, err)
	}
}

func TestExpandInstallModuleNames(t *testing.T) {
	module.DiscoveredAddons = map[string]*module.Addon{
		"base": {Manifest: module.Manifest{Name: "base", Depends: []string{}}},
		"sale": {Manifest: module.Manifest{Name: "sale", Depends: []string{"base"}}},
	}
	t.Cleanup(func() { module.DiscoveredAddons = nil })
	got, err := module.ExpandInstallModuleNamesForTest(t.Context(), []string{"sale"})
	if err != nil || len(got) != 1 || got[0] != "sale" {
		t.Fatalf("expand: %v err=%v", got, err)
	}
}

func TestDataFileOptsSkipExisting(t *testing.T) {
	ctx := module.ContextWithSyncMode(t.Context(), module.ModuleReloadUpdateForTest)
	opts := module.NewDataFileOptsForTest(true)
	if opts.SkipExistingOnUpdateForTest(ctx, "base", "") {
		t.Fatal("empty xml id should not skip")
	}
	opts2 := module.NewDataFileOptsForTest(false)
	if opts2.SkipExistingOnUpdateForTest(ctx, "base", "xmlid") {
		t.Fatal("update without noUpdate should not skip")
	}
}

func TestModuleReloadConstants(t *testing.T) {
	if module.ModuleReloadInstallForTest == module.ModuleReloadUpdateForTest {
		t.Fatal("reload constants should differ")
	}
}

func TestDiscoverTempAddon(t *testing.T) {
	dir := t.TempDir()
	harness.WriteTempGoMod(t, dir)
	base := harness.WriteMinimalBaseAddon(t, dir)
	discovered, err := module.DiscoverAddonRoots([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := discovered["base"]; !ok {
		t.Fatalf("discovered=%v path=%s", discovered, base.Path)
	}
	abs := filepath.Clean(base.Path)
	if discovered["base"].Path != abs {
		t.Fatalf("path=%q want %q", discovered["base"].Path, abs)
	}
}
