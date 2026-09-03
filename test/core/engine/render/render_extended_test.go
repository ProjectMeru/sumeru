package render_test

import (
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)

func TestBuildSidebarMenusExtended(t *testing.T) {
	menus := []parser.MenuItem{
		{ID: "1", Name: "Root", ParentID: ""},
		{ID: "2", Name: "Sales", ParentID: "1"},
		{ID: "3", Name: "Leads", ParentID: "2"},
	}
	out := render.BuildSidebarMenus(menus, "1", func(m parser.MenuItem) bool { return true })
	if len(out) != 1 || out[0].Name != "Sales" {
		t.Fatalf("sidebar: %+v", out)
	}
}

func TestResolveActiveModuleID(t *testing.T) {
	menus := []parser.MenuItem{
		{ID: "10", Name: "CRM", ParentID: ""},
		{ID: "11", Name: "Leads", ParentID: "10"},
	}
	if got := render.ResolveActiveModuleID(menus, "11"); got != "10" {
		t.Fatalf("module id: %q", got)
	}
}

func TestViewModeFilterSetExtended(t *testing.T) {
	set := render.ViewModeFilterSetForTest([]string{"list", "form"})
	if _, ok := set["list"]; !ok {
		t.Fatalf("missing list in %v", set)
	}
	if _, ok := set["form"]; !ok {
		t.Fatalf("missing form in %v", set)
	}
}

func TestWorkspaceRecordBreadcrumbLabelListExtended(t *testing.T) {
	label := render.WorkspaceRecordBreadcrumbLabelForTest(render.BreadcrumbInput{
		ViewType: render.ViewModeKanban,
		ResModel: "crm.lead",
	})
	if label != "" {
		t.Fatalf("kanban label should be empty, got %q", label)
	}
}
