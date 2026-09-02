package web

import (
	"net/http/httptest"
	"testing"
)

func TestParseWorkspaceRequest(t *testing.T) {
	req := httptest.NewRequest("GET", workspaceRoute+"?menu_id=5&view_type=form&id=42&edit=1", nil)
	got := parseWorkspaceRequest(req, 12)
	if got.actionID != 12 || got.menuID != "5" || got.viewType != "form" || got.recordID != "42" || !got.formEdit {
		t.Fatalf("unexpected request: %+v", got)
	}
}

func TestHomeRouteWithMenu(t *testing.T) {
	if got := homeRouteWithMenu(""); got != homeRoute {
		t.Fatalf("empty menu should return home, got %q", got)
	}
	if got := homeRouteWithMenu(" 7 "); got != homeRoute+"?menu_id=7" {
		t.Fatalf("unexpected URL: %q", got)
	}
}

func TestPrependViewMode(t *testing.T) {
	modes := prependViewMode(workspaceViewModeForm, []string{workspaceViewModeList})
	if len(modes) != 2 || modes[0] != workspaceViewModeForm || modes[1] != workspaceViewModeList {
		t.Fatalf("unexpected modes: %v", modes)
	}
	if got := prependViewMode("", []string{workspaceViewModeList}); len(got) != 1 {
		t.Fatalf("empty mode should not prepend, got %v", got)
	}
}

func TestIsNumericRecordID(t *testing.T) {
	if !isNumericRecordID("42") {
		t.Fatal("42 should be numeric")
	}
	if isNumericRecordID("abc") {
		t.Fatal("abc should not be numeric")
	}
}

func TestWorkspaceViewModeCandidates(t *testing.T) {
	req := httptest.NewRequest("GET", workspaceRoute+"?view_type=Kanban&id=9", nil)
	actionData := map[string]interface{}{"view_mode": "list,form"}

	modes := workspaceViewModeCandidates(req, actionData)
	want := []string{workspaceViewModeForm, "kanban", workspaceViewModeList, workspaceViewModeForm}
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
	if got := actionViewModesForTabs(nil); got != nil {
		t.Fatalf("nil action should not filter tabs, got %v", got)
	}
	if got := actionViewModesForTabs(map[string]interface{}{}); got != nil {
		t.Fatalf("empty action should not filter tabs, got %v", got)
	}
	if got := actionViewModesForTabs(map[string]interface{}{"view_mode": "map,list,form"}); len(got) != 3 || got[0] != "map" {
		t.Fatalf("got %v", got)
	}
	if got := actionViewModesForTabs(map[string]interface{}{"view_mode": ""}); len(got) != 1 || got[0] != workspaceViewModeList {
		t.Fatalf("empty view_mode should default to list, got %v", got)
	}
}

func TestParsePositiveRecordID(t *testing.T) {
	recordID, ok := parsePositiveRecordID("15")
	if !ok || recordID != 15 {
		t.Fatalf("got id=%d ok=%v", recordID, ok)
	}
	if _, ok := parsePositiveRecordID("0"); ok {
		t.Fatal("zero id should be rejected")
	}
	if _, ok := parsePositiveRecordID("bad"); ok {
		t.Fatal("non-numeric id should be rejected")
	}
}
