package module_test

import (
	"strings"
	"testing"

	"sumeru/core/module"
)

func TestValidateDiscoveredDependsMissing(t *testing.T) {
	discovered := map[string]*module.Addon{
		"demo": {Manifest: module.Manifest{Name: "demo", Depends: []string{"missing_mod"}}},
	}
	err := module.ValidateDiscoveredDepends(discovered)
	if err == nil {
		t.Fatal("expected error for missing depend")
	}
	if !strings.Contains(err.Error(), "missing_mod") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveInstallClosureEngagementCookbook(t *testing.T) {
	discovered := map[string]*module.Addon{
		"base": {Manifest: module.Manifest{Name: "base", Depends: nil}},
		"contacts": {Manifest: module.Manifest{Name: "contacts", Depends: []string{"base"}}},
		"mail": {Manifest: module.Manifest{Name: "mail", Depends: []string{"base"}}},
		"hr": {Manifest: module.Manifest{Name: "hr", Depends: []string{"base", "contacts"}}},
		"engagement_cookbook": {
			Manifest: module.Manifest{
				Name:    "engagement_cookbook",
				Depends: []string{"base", "contacts", "hr", "mail"},
			},
		},
	}
	closure, err := module.ResolveInstallClosure(discovered, "engagement_cookbook")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"base", "contacts", "hr", "mail", "engagement_cookbook"}
	if len(closure) != len(want) {
		t.Fatalf("closure len = %d, want %d: %v", len(closure), len(want), closure)
	}
	index := map[string]int{}
	for i, name := range closure {
		index[name] = i
	}
	for _, name := range want {
		if _, ok := index[name]; !ok {
			t.Fatalf("missing %q in closure %v", name, closure)
		}
	}
	if index["base"] >= index["hr"] || index["hr"] >= index["engagement_cookbook"] {
		t.Fatalf("unexpected order: %v", closure)
	}
}

func TestResolveInstallClosureMissingDepend(t *testing.T) {
	discovered := map[string]*module.Addon{
		"demo": {Manifest: module.Manifest{Name: "demo", Depends: []string{"hr"}}},
	}
	_, err := module.ResolveInstallClosure(discovered, "demo")
	if err == nil || !strings.Contains(err.Error(), "not on addons_path") {
		t.Fatalf("expected missing depend error, got %v", err)
	}
}
