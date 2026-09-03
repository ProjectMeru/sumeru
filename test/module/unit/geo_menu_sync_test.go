package module_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/module"
)

func baseAddonDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("..", "..", "..", "addons", "base")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("base addon not found at ", dir)
	}
	return dir
}

func collectBaseManifestMenus(t *testing.T) []parser.MenuItem {
	t.Helper()
	baseDir := baseAddonDir(t)
	manifestPath := filepath.Join(baseDir, "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	var out []parser.MenuItem
	for _, rel := range manifest.Data {
		if !strings.HasSuffix(strings.ToLower(rel), ".xml") {
			continue
		}
		items, err := module.CollectMenuItemsFromManifestFile(filepath.Join(baseDir, rel))
		if err != nil {
			t.Fatalf("collect menus from %s: %v", rel, err)
		}
		out = append(out, items...)
	}
	return out
}

func menuByID(menus []parser.MenuItem, id string) (parser.MenuItem, bool) {
	for _, m := range menus {
		if m.ID == id {
			return m, true
		}
	}
	return parser.MenuItem{}, false
}

func TestBaseDeferredMenusIncludeGeoSection(t *testing.T) {
	menus := collectBaseManifestMenus(t)
	geoSection, ok := menuByID(menus, "menu_geo_section")
	if !ok {
		t.Fatal("menu_geo_section missing from collected base manifest menus")
	}
	if geoSection.ParentID != "menu_settings_root" {
		t.Fatalf("menu_geo_section parent = %q; want menu_settings_root", geoSection.ParentID)
	}
	if _, ok := menuByID(menus, "menu_settings_root"); !ok {
		t.Fatal("menu_settings_root missing — geo section parent would be unresolved")
	}
	for _, childID := range []string{"menu_core_country", "menu_core_country_state", "menu_core_city"} {
		child, ok := menuByID(menus, childID)
		if !ok {
			t.Fatalf("%s missing from collected menus", childID)
		}
		if child.ParentID != "menu_geo_section" {
			t.Fatalf("%s parent = %q; want menu_geo_section", childID, child.ParentID)
		}
		if !strings.Contains(child.AccessGroups, "base.group_system") {
			t.Fatalf("%s access_groups = %q; want base.group_system", childID, child.AccessGroups)
		}
	}
}

func TestBaseManifestMenusLoadLast(t *testing.T) {
	baseDir := baseAddonDir(t)
	raw, err := os.ReadFile(filepath.Join(baseDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Data) == 0 {
		t.Fatal("manifest data is empty")
	}
	last := manifest.Data[len(manifest.Data)-1]
	if last != "views/menus.xml" {
		t.Fatalf("views/menus.xml must be last in manifest data; got %q", last)
	}
}
