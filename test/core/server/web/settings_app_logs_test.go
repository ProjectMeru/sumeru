package web_test

import (
	"sumeru/core/server/web"
	"testing"
)

func TestAppLogsViewStylesheets(t *testing.T) {
	stylesheets := web.AppLogsViewStylesheets()
	if len(stylesheets) != 2 || stylesheets[0] != web.TestWorkspaceStylesheetURL || stylesheets[1] != web.TestPagesStylesheetURL {
		t.Fatalf("unexpected stylesheets: %v", stylesheets)
	}
}

func TestBuildAppLogsPageDataWithoutMenu(t *testing.T) {
	page := web.BuildAppLogsPageData(t.Context(), 0)
	if page.Title != web.TestAppLogsPageTitle || page.ViewBreadcrumb != web.TestAppLogsBreadcrumb || !page.SettingsNavActive {
		t.Fatalf("unexpected page data: %+v", page)
	}
	if len(page.BreadcrumbItems) != 0 {
		t.Fatal("expected no breadcrumbs without menu id")
	}
}

func TestResolveMenuIDMissing(t *testing.T) {
	menuID, menuIDStr := web.ResolveMenuID(t.Context(), "missing.menu_xml_id")
	if menuID != 0 || menuIDStr != "" {
		t.Fatalf("got menuID=%d menuIDStr=%q", menuID, menuIDStr)
	}
}
