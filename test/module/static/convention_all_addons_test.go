package static_test

import (
	"os"
	"path/filepath"
	"testing"

	"sumeru/core/module"
	"sumeru/test/harness"
)

func TestDiscoveredCoreAddonsPassConvention(t *testing.T) {
	root := harness.RepoRoot(t)
	addonsPath := filepath.Join(root, "addons")
	discovered, err := module.DiscoverAddonRoots([]string{addonsPath})
	if err != nil {
		t.Fatal(err)
	}
	if len(discovered) == 0 {
		t.Fatal("expected core addons under addons/")
	}
	if err := module.ValidateDiscoveredAddons(discovered); err != nil {
		t.Fatalf("core addons convention: %v", err)
	}
}

func TestDiscoveredExampleConfigAddonsPassConvention(t *testing.T) {
	discovered := harness.DiscoverFromExampleConfig(t)
	if err := module.ValidateDiscoveredAddons(discovered); err != nil {
		t.Fatalf("sumeru.conf.example addons convention: %v", err)
	}
}

func TestResolveInstallClosureCoreAddons(t *testing.T) {
	root := harness.RepoRoot(t)
	addonsPath := filepath.Join(root, "addons")
	discovered, err := module.DiscoverAddonRoots([]string{addonsPath})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"base", "contacts", "mail"} {
		if _, ok := discovered[name]; !ok {
			continue
		}
		closure, err := module.ResolveInstallClosure(discovered, name)
		if err != nil {
			t.Fatalf("ResolveInstallClosure(%q): %v", name, err)
		}
		if len(closure) == 0 {
			t.Fatalf("empty closure for %q", name)
		}
		if closure[0] != "base" && name != "base" {
			t.Fatalf("closure for %q should start with base, got %v", name, closure)
		}
	}
}

func TestWorkspaceAddonsPathOptional(t *testing.T) {
	root := harness.RepoRoot(t)
	sibling := filepath.Join(filepath.Dir(root), "sumeru_addons")
	if _, err := os.Stat(sibling); err != nil {
		t.Skipf("sumeru_addons not present at %s", sibling)
	}
	discovered, err := module.DiscoverAddonRoots([]string{filepath.Join(root, "addons"), sibling})
	if err != nil {
		t.Fatal(err)
	}
	if err := module.ValidateDiscoveredAddons(discovered); err != nil {
		t.Fatalf("core + sumeru_addons convention: %v", err)
	}
}
