package render_test

import (
	"testing"

	"sumeru/core/engine/render"
)



func TestSplitModuleMenuChain(t *testing.T) {
	chain := []render.MenuCrumbForTest{
		{ID: 1, Name: "CRM"},
		{ID: 2, Name: "Sales"},
		{ID: 3, Name: "Pipeline", ActionID: 9},
	}
	mod, menus, ok := render.SplitModuleMenuChainForTest(chain)
	if !ok || mod.Name != "CRM" || len(menus) != 2 || menus[1].Name != "Pipeline" {
		t.Fatalf("split: mod=%+v menus=%+v ok=%v", mod, menus, ok)
	}
}

func TestWorkspaceViewBreadcrumbLabel(t *testing.T) {
	if got := render.WorkspaceViewBreadcrumbLabelForTest(render.ViewModeKanban); got != "" {
		t.Fatalf("kanban view label should be empty, got %q", got)
	}
	if got := render.WorkspaceViewBreadcrumbLabelForTest(render.ViewModeGraph); got != "Graph" {
		t.Fatalf("graph = %q, want Graph", got)
	}
}

func TestWorkspaceRecordBreadcrumbLabel_newFormUsesModelFallback(t *testing.T) {
	label := render.WorkspaceRecordBreadcrumbLabelForTest(render.BreadcrumbInput{
		ResModel: "crm.lead",
		ViewType: render.ViewModeForm,
		RecordID: 0,
	})
	if label != "Lead" {
		t.Fatalf("new form crm.lead = %q, want Lead (generic fallback)", label)
	}
}

func TestWorkspaceRecordBreadcrumbLabel_existingRecord(t *testing.T) {
	label := render.WorkspaceRecordBreadcrumbLabelForTest(render.BreadcrumbInput{
		ViewType: render.ViewModeForm,
		RecordID: 7,
		Record:   map[string]interface{}{"name": "Acme Deal"},
	})
	if label != "Acme Deal" {
		t.Fatalf("record name = %q, want Acme Deal", label)
	}
}

func TestUIModelName_genericAddonModel(t *testing.T) {
	if got := render.UIModelName("sale.order"); got != "Order" {
		t.Fatalf("sale.order = %q, want Order", got)
	}
}
