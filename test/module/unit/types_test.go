package module_test

import (
	"testing"

	"sumeru/core/module"
)

func TestManifestIsAutoInstall(t *testing.T) {
	m := module.Manifest{Name: "sale_crm"}
	if m.IsAutoInstall() {
		t.Fatal("default auto_install should be false")
	}
	auto := true
	m.AutoInstall = &auto
	if !m.IsAutoInstall() {
		t.Fatal("expected auto_install true")
	}
}

func TestSortDiscoveredAddonsTopoSaleCRM(t *testing.T) {
	discovered := map[string]*module.Addon{
		"sale_crm": {Manifest: module.Manifest{Name: "sale_crm", Depends: []string{"crm", "sale"}, AutoInstall: boolPtr(true)}},
		"sale":     {Manifest: module.Manifest{Name: "sale", Depends: []string{"base"}}},
		"crm":      {Manifest: module.Manifest{Name: "crm", Depends: []string{"base"}}},
		"base":     {Manifest: module.Manifest{Name: "base"}},
	}
	topo, err := module.SortDiscoveredAddonsTopo(discovered)
	if err != nil {
		t.Fatal(err)
	}
	idx := map[string]int{}
	for i, a := range topo {
		idx[a.Manifest.Name] = i
	}
	if idx["crm"] >= idx["sale_crm"] || idx["sale"] >= idx["sale_crm"] {
		t.Fatalf("sale_crm should come after crm and sale: %#v", topo)
	}
}

func boolPtr(v bool) *bool { return &v }
