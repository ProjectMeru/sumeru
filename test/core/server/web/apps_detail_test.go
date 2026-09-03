package web_test

import (
	"net/url"
	"sumeru/core/server/web"
	"testing"
)

func TestAppsDetailURL(t *testing.T) {
	assertQueryContains(t, web.AppsDetailURL(web.AppsBrowseState{
		Layout: "grid", Filter: "installed", Scope: "apps", SearchQuery: "crm", ModuleName: "sale",
	}, true), map[string]string{
		"layout": "grid",
		"filter": "installed",
		"scope":  "apps",
		"q":      "crm",
		"module": "sale",
		"edit":   "1",
	})

	assertQueryContains(t, web.AppsDetailURL(web.AppsBrowseState{
		Layout: "list", Filter: "all", Scope: "all", ModuleName: "base",
	}, false), map[string]string{
		"layout": "list",
		"module": "base",
	})
}

func assertQueryContains(t *testing.T, rawURL string, want map[string]string) {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse %q: %v", rawURL, err)
	}
	got := parsed.Query()
	for key, value := range want {
		if got.Get(key) != value {
			t.Fatalf("url %q: got %s=%q want %q", rawURL, key, got.Get(key), value)
		}
	}
	for key := range got {
		if _, ok := want[key]; !ok && got.Get(key) != "" {
			t.Fatalf("url %q: unexpected param %s=%q", rawURL, key, got.Get(key))
		}
	}
}

func TestFindAppsModule(t *testing.T) {
	modules := []web.AppsModule{
		{Name: "sale", CanInstall: false},
		{Name: "crm", CanInstall: true},
	}
	found, ok := web.FindAppsModule(modules, "crm")
	if !ok || !found.CanInstall {
		t.Fatalf("expected crm entry, got %+v ok=%v", found, ok)
	}
	_, ok = web.FindAppsModule(modules, "missing")
	if ok {
		t.Fatal("expected false for missing module")
	}
}
