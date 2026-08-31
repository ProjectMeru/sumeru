package web_test

import (
	"net/http/httptest"
	"sumeru/core/server/web"
	"testing"
)

func TestParseWorkspaceRequest(t *testing.T) {
	req := httptest.NewRequest("GET", web.TestWorkspaceRoute+"?menu_id=5&view_type=form&id=42&edit=1", nil)
	got := web.ParseWorkspaceRequest(req, 12)
	actionID, menuID, viewType, recordID, formEdit := web.WorkspaceRequestFields(got)

	if actionID != 12 || menuID != "5" || viewType != "form" || recordID != "42" || !formEdit {
		t.Fatalf("unexpected request: id=%d menu=%q view=%q record=%q edit=%v", actionID, menuID, viewType, recordID, formEdit)
	}
}

func TestHomeRouteWithMenu(t *testing.T) {
	if got := web.HomeRouteWithMenu(""); got != web.TestHomeRoute {
		t.Fatalf("empty menu should return home, got %q", got)
	}
	if got := web.HomeRouteWithMenu(" 7 "); got != web.TestHomeRoute+"?menu_id=7" {
		t.Fatalf("unexpected URL: %q", got)
	}
}

func TestPrependViewMode(t *testing.T) {
	modes := web.PrependViewMode(web.TestWorkspaceViewModeForm, []string{web.TestWorkspaceViewModeList})
	if len(modes) != 2 || modes[0] != web.TestWorkspaceViewModeForm || modes[1] != web.TestWorkspaceViewModeList {
		t.Fatalf("unexpected modes: %v", modes)
	}
	if got := web.PrependViewMode("", []string{web.TestWorkspaceViewModeList}); len(got) != 1 {
		t.Fatalf("empty mode should not prepend, got %v", got)
	}
}

func TestIsNumericRecordID(t *testing.T) {
	if !web.IsNumericRecordID("42") {
		t.Fatal("42 should be numeric")
	}
	if web.IsNumericRecordID("abc") {
		t.Fatal("abc should not be numeric")
	}
}

func TestWorkspaceViewModeCandidates(t *testing.T) {
	req := httptest.NewRequest("GET", web.TestWorkspaceRoute+"?view_type=Kanban&id=9", nil)
	actionData := map[string]interface{}{"view_mode": "list,form"}

	modes := web.WorkspaceViewModeCandidates(req, actionData)
	want := []string{web.TestWorkspaceViewModeForm, "kanban", web.TestWorkspaceViewModeList, web.TestWorkspaceViewModeForm}
	if len(modes) != len(want) {
		t.Fatalf("modes = %v, want %v", modes, want)
	}
	for i := range want {
		if modes[i] != want[i] {
			t.Fatalf("modes[%d] = %q, want %q (full: %v)", i, modes[i], want[i], modes)
		}
	}
}

func TestActionViewModesForTabs(t *testing.T) {
	if got := web.ActionViewModesForTabs(nil); got != nil {
		t.Fatalf("nil action should not filter tabs, got %v", got)
	}
	if got := web.ActionViewModesForTabs(map[string]interface{}{}); got != nil {
		t.Fatalf("empty action should not filter tabs, got %v", got)
	}
	if got := web.ActionViewModesForTabs(map[string]interface{}{"view_mode": "map,list,form"}); len(got) != 3 || got[0] != "map" {
		t.Fatalf("got %v", got)
	}
	if got := web.ActionViewModesForTabs(map[string]interface{}{"view_mode": ""}); len(got) != 1 || got[0] != web.TestWorkspaceViewModeList {
		t.Fatalf("empty view_mode should default to list, got %v", got)
	}
}

func TestParsePositiveRecordID(t *testing.T) {
	recordID, ok := web.ParsePositiveRecordID("15")
	if !ok || recordID != 15 {
		t.Fatalf("got id=%d ok=%v", recordID, ok)
	}
	if _, ok := web.ParsePositiveRecordID("0"); ok {
		t.Fatal("zero id should be rejected")
	}
	if _, ok := web.ParsePositiveRecordID("bad"); ok {
		t.Fatal("non-numeric id should be rejected")
	}
}
