package module_test

import (
	"testing"

	"sumeru/core/module"
)

func TestSortDiscoveredAddonsTopoDependsFirst(t *testing.T) {
	discovered := map[string]*module.Addon{
		"crm_iap_enrich": {Manifest: module.Manifest{Name: "crm_iap_enrich", Depends: []string{"iap", "crm"}}},
		"crm":            {Manifest: module.Manifest{Name: "crm", Depends: []string{"base"}}},
		"iap":            {Manifest: module.Manifest{Name: "iap", Depends: []string{"base"}}},
		"base":           {Manifest: module.Manifest{Name: "base", Depends: []string{}}},
	}
	topo, err := module.SortDiscoveredAddonsTopo(discovered)
	if err != nil {
		t.Fatal(err)
	}
	index := map[string]int{}
	for i, addon := range topo {
		index[addon.Manifest.Name] = i
	}
	if index["base"] >= index["crm"] || index["crm"] >= index["crm_iap_enrich"] {
		t.Fatalf("unexpected topo order: %#v", topo)
	}
}
