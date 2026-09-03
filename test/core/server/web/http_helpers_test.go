package web_test

import (
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestSplitCommaSeparatedValues(t *testing.T) {
	got := web.SplitCommaSeparatedValues(" list, form ,,kanban ")
	want := []string{"list", "form", "kanban"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestFirstGroupByField(t *testing.T) {
	if got := web.FirstGroupByField(" state, user_id "); got != "state" {
		t.Fatalf("got %q want state", got)
	}
	if got := web.FirstGroupByField(""); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestSplitViewModesDelegatesToCommaSplit(t *testing.T) {
	raw := "tree,form"
	if strings.Join(web.SplitViewModes(raw), ",") != strings.Join(web.SplitCommaSeparatedValues(raw), ",") {
		t.Fatal("web.SplitViewModes should match web.SplitCommaSeparatedValues")
	}
}

func TestNormalizeViewMode(t *testing.T) {
	if got := web.NormalizeViewMode(" Form "); got != "form" {
		t.Fatalf("got %q want %q", got, "form")
	}
}

func TestFormBaseQueryValues(t *testing.T) {
	got := web.FormBaseQueryValues(12, " 3 ", "form", " 99 ")
	assertQueryContains(t, "/web?"+got, map[string]string{
		web.TestWorkspaceActionParam:   "12",
		web.TestWorkspaceMenuIDParam:   "3",
		web.TestWorkspaceViewTypeParam: "form",
		web.TestWorkspaceRecordIDParam: "99",
	})

	if empty := web.FormBaseQueryValues(0, "", "", ""); empty != "" {
		t.Fatalf("got %q want empty query", empty)
	}
}

func TestWorkspaceListURL(t *testing.T) {
	got := web.WorkspaceListURL("5", "2")
	assertQueryContains(t, got, map[string]string{
		web.TestWorkspaceActionParam: "5",
		web.TestWorkspaceMenuIDParam: "2",
	})
	if empty := web.WorkspaceListURL("", ""); empty != web.TestWorkspaceRoute {
		t.Fatalf("got %q want %q", empty, web.TestWorkspaceRoute)
	}
}

func TestFormOrQueryValue(t *testing.T) {
	req := httptest.NewRequest("POST", "/web/record/delete?model=fallback", strings.NewReader("model=core.user"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	if got := web.FormOrQueryValue(req, web.TestRecordModelField); got != "core.user" {
		t.Fatalf("form value preferred: got %q", got)
	}

	req = httptest.NewRequest("GET", "/web/record/delete?model=core.company", nil)
	if got := web.FormOrQueryValue(req, web.TestRecordModelField); got != "core.company" {
		t.Fatalf("query fallback: got %q", got)
	}
}
