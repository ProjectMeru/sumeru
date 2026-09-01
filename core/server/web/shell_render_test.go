package web_test

import (
	"net/http/httptest"
	"reflect"
	"sumeru/core/server/web"
	"testing"

	"sumeru/core/engine/render"
)

func TestResolveShellRoute(t *testing.T) {
	req := httptest.NewRequest("GET", web.TestAppLogsRoute, nil)
	if got := web.ResolveShellRoute(web.ShellPageOpts{Route: web.TestSettingsRoute}, req); got != web.TestSettingsRoute {
		t.Fatalf("got %q want explicit route", got)
	}
	if got := web.ResolveShellRoute(web.ShellPageOpts{}, req); got != web.TestAppLogsRoute {
		t.Fatalf("got %q want request path", got)
	}
}

func TestResolveExtraStylesheets(t *testing.T) {
	optStylesheets := []string{"/static/css/custom.css"}
	if got := web.ResolveExtraStylesheets(nil, optStylesheets); !reflect.DeepEqual(got, optStylesheets) {
		t.Fatalf("opts override: got %v", got)
	}

	pageStylesheets := []string{"/static/css/page.css"}
	if got := web.ResolveExtraStylesheets(pageStylesheets, nil); !reflect.DeepEqual(got, pageStylesheets) {
		t.Fatalf("page stylesheets: got %v", got)
	}

	if got := web.ResolveExtraStylesheets(nil, nil); !reflect.DeepEqual(got, render.ExtraStylesheetURLs) {
		t.Fatalf("default stylesheets: got %v", got)
	}
}

func TestResolveExtraScripts(t *testing.T) {
	optScripts := []string{"/static/addon-asset/demo/static/js/widget.js"}
	if got := web.ResolveExtraScripts(nil, optScripts); !reflect.DeepEqual(got, optScripts) {
		t.Fatalf("opts override: got %v", got)
	}

	pageScripts := []string{"/static/js/page.js"}
	if got := web.ResolveExtraScripts(pageScripts, nil); !reflect.DeepEqual(got, pageScripts) {
		t.Fatalf("page scripts: got %v", got)
	}

	if got := web.ResolveExtraScripts(nil, nil); !reflect.DeepEqual(got, render.ExtraScriptURLs) {
		t.Fatalf("default scripts: got %v", got)
	}
}

func TestApplyShellPageDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", web.TestHomeRoute, nil)
	page := web.ApplyShellPageDefaults(render.PageData{}, web.ShellPageOpts{MenuIDStr: "12"}, web.TestHomeRoute, req, "sale", nil)

	if page.Title != web.TestDefaultPageTitle {
		t.Fatalf("title=%q want %q", page.Title, web.TestDefaultPageTitle)
	}
	if page.ModuleName != "sale" || page.ActiveMenuID != "12" || !page.HomeNavActive {
		t.Fatalf("unexpected page defaults: %+v", page)
	}
	if !page.SuppressSidebar {
		t.Fatal("empty sidebar should suppress sidebar")
	}
	if len(page.ViewStylesheetURLs) != 1 || page.ViewStylesheetURLs[0] != web.TestWorkspaceStylesheetURL {
		t.Fatalf("unexpected stylesheets: %v", page.ViewStylesheetURLs)
	}
}
