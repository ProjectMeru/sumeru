package render

import "testing"

func TestSplitModuleMenuChain(t *testing.T) {
	chain := []menuCrumb{
		{ID: 1, Name: "CRM"},
		{ID: 2, Name: "Sales"},
		{ID: 3, Name: "Pipeline", ActionID: 9},
	}
	mod, menus, ok := splitModuleMenuChain(chain)
	if !ok || mod.Name != "CRM" || len(menus) != 2 || menus[1].Name != "Pipeline" {
		t.Fatalf("split: mod=%+v menus=%+v ok=%v", mod, menus, ok)
	}
}

func TestWorkspaceViewBreadcrumbLabel(t *testing.T) {
	if got := workspaceViewBreadcrumbLabel(ViewModeKanban); got != "" {
		t.Fatalf("kanban view label should be empty, got %q", got)
	}
	if got := workspaceViewBreadcrumbLabel(ViewModeGraph); got != "Graph" {
		t.Fatalf("graph = %q, want Graph", got)
	}
}

func TestWorkspaceRecordBreadcrumbLabel_newFormUsesModelFallback(t *testing.T) {
	label := workspaceRecordBreadcrumbLabel(BreadcrumbInput{
		ResModel: "crm.lead",
		ViewType: ViewModeForm,
		RecordID: 0,
	})
	if label != "Lead" {
		t.Fatalf("new form crm.lead = %q, want Lead (generic fallback)", label)
	}
}

func TestWorkspaceRecordBreadcrumbLabel_existingRecord(t *testing.T) {
	label := workspaceRecordBreadcrumbLabel(BreadcrumbInput{
		ViewType: ViewModeForm,
		RecordID: 7,
		Record:   map[string]interface{}{"name": "Acme Deal"},
	})
	if label != "Acme Deal" {
		t.Fatalf("record name = %q, want Acme Deal", label)
	}
}

func TestUIModelName_genericAddonModel(t *testing.T) {
	if got := UIModelName("sale.order"); got != "Order" {
		t.Fatalf("sale.order = %q, want Order", got)
	}
}
