package render_test

import (
	"testing"
	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)


func TestBuildSidebarMenus_localizationSection(t *testing.T) {
	allMenus := []parser.MenuItem{
		{ID: "100", Name: "Settings", Sequence: 5},
		{ID: "160", Name: "Localization", ParentID: "100", Sequence: 60, AccessGroups: "base.group_system"},
		{ID: "161", Name: "Countries", ParentID: "160", Sequence: 10, Action: "/web?action=10&menu_id=161", AccessGroups: "base.group_system"},
		{ID: "162", Name: "States", ParentID: "160", Sequence: 20, Action: "/web?action=11&menu_id=162", AccessGroups: "base.group_system"},
		{ID: "163", Name: "Cities", ParentID: "160", Sequence: 30, Action: "/web?action=12&menu_id=163", AccessGroups: "base.group_system"},
	}
	menuAllowed := func(mi parser.MenuItem) bool {
		return mi.AccessGroups == "base.group_system"
	}
	sections := render.BuildSidebarMenus(allMenus, "100", menuAllowed)
	var localization *render.SidebarMenu
	for i := range sections {
		if sections[i].Name == "Localization" {
			localization = &sections[i]
			break
		}
	}
	if localization == nil {
		t.Fatal("Localization section not found in Settings sidebar")
	}
	if len(localization.SubMenus) != 3 {
		t.Fatalf("Localization submenus = %d; want 3", len(localization.SubMenus))
	}
	names := map[string]bool{}
	for _, sm := range localization.SubMenus {
		names[sm.Name] = true
		if sm.Action == "" {
			t.Fatalf("submenu %q has empty action URL", sm.Name)
		}
	}
	for _, want := range []string{"Countries", "States", "Cities"} {
		if !names[want] {
			t.Fatalf("missing submenu %q", want)
		}
	}
}

func TestBuildSidebarMenus_skipsEmptySections(t *testing.T) {
	allMenus := []parser.MenuItem{
		{ID: "200", Name: "Contacts", Sequence: 20},
		{ID: "210", Name: "Contacts organization", ParentID: "200", Sequence: 100},
		{ID: "211", Name: "All Contacts", ParentID: "210", Sequence: 10, Action: "/web?action=1&menu_id=211"},
		{ID: "220", Name: "Empty section", ParentID: "200", Sequence: 200},
	}
	menuAllowed := func(parser.MenuItem) bool { return true }
	sections := render.BuildSidebarMenus(allMenus, "200", menuAllowed)
	if len(sections) != 1 {
		t.Fatalf("sections = %d; want 1 (empty section skipped)", len(sections))
	}
	if sections[0].Name != "Contacts organization" {
		t.Fatalf("section name = %q; want Contacts organization", sections[0].Name)
	}
}

func TestSidebarHasMenus(t *testing.T) {
	if render.SidebarHasMenus(nil) {
		t.Fatal("nil menus should be false")
	}
	if render.SidebarHasMenus([]render.SidebarMenu{{Name: "Empty", SubMenus: nil}}) {
		t.Fatal("section without links should be false")
	}
	if !render.SidebarHasMenus([]render.SidebarMenu{{Name: "Links", SubMenus: []parser.MenuItem{{Name: "One"}}}}) {
		t.Fatal("section with links should be true")
	}
}

func TestResolveActiveModuleID_noDefaultWhenMenuMissing(t *testing.T) {
	got := render.ResolveActiveModuleID(nil, "")
	if got != "" {
		t.Fatalf("active module = %q; want empty when menu_id missing", got)
	}
}
