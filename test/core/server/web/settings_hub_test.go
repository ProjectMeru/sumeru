package web_test

import (
	"sumeru/core/server/web"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)

func TestSettingsHubSectionFromSidebar(t *testing.T) {
	section, ok := web.SettingsHubSectionFromSidebar(render.SidebarMenu{
		Name: "Companies",
		SubMenus: []parser.MenuItem{
			{Name: "All Companies", Action: "/web?menu_id=10"},
			{Name: "", Action: "/web?menu_id=11"},
		},
	})
	if !ok {
		t.Fatal("expected section with valid links")
	}
	if section.Title != "Companies" || len(section.Links) != 1 {
		t.Fatalf("unexpected section: %+v", section)
	}
	if section.FilterText != "companies all companies" {
		t.Fatalf("unexpected filter text: %q", section.FilterText)
	}
}

func TestSettingsHubSectionFromSidebarEmpty(t *testing.T) {
	_, ok := web.SettingsHubSectionFromSidebar(render.SidebarMenu{
		Name: "Users",
		SubMenus: []parser.MenuItem{
			{Name: "Users", Action: ""},
		},
	})
	if ok {
		t.Fatal("section without actionable links should be skipped")
	}
}

func TestBuildSettingsHubPageData(t *testing.T) {
	page := web.BuildSettingsHubPageData(t.Context(), "5")
	if page.Title != web.TestSettingsHubPageTitle || !page.SettingsNavActive || page.ActiveMenuID != "5" {
		t.Fatalf("unexpected page data: %+v", page)
	}
	if len(page.ViewStylesheetURLs) != 1 || page.ViewStylesheetURLs[0] != web.TestSettingsHubStylesheetURL {
		t.Fatalf("unexpected stylesheets: %v", page.ViewStylesheetURLs)
	}
}
