package web_test

import (
	"net/http/httptest"
	"sumeru/core/server/web"
	"testing"
)

func TestNormalizeGridListLayout(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "", want: web.TestAppsLayoutGrid},
		{raw: web.TestAppsLayoutGrid, want: web.TestAppsLayoutGrid},
		{raw: web.TestAppsLayoutList, want: web.TestAppsLayoutList},
		{raw: web.TestLegacyKanbanLayout, want: web.TestAppsLayoutGrid},
		{raw: "GRID", want: web.TestAppsLayoutGrid},
		{raw: " List ", want: web.TestAppsLayoutList},
		{raw: "table", want: web.TestAppsLayoutGrid},
	}
	for _, test := range tests {
		if got := web.NormalizeGridListLayout(test.raw); got != test.want {
			t.Errorf("web.NormalizeGridListLayout(%q) = %q, want %q", test.raw, got, test.want)
		}
	}
}

func TestLayoutFromQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/web/apps?layout=list", nil)
	if got := web.LayoutFromQuery(req); got != web.TestAppsLayoutList {
		t.Fatalf("got %q want %q", got, web.TestAppsLayoutList)
	}
}

func TestLayoutFromForm(t *testing.T) {
	req := httptest.NewRequest("POST", "/web/module/action", nil)
	req.Form = map[string][]string{web.TestAppsLayoutField: {web.TestLegacyKanbanLayout}}
	if got := web.LayoutFromForm(req, web.TestAppsLayoutField); got != web.TestAppsLayoutGrid {
		t.Fatalf("legacy kanban should map to grid, got %q", got)
	}
}
