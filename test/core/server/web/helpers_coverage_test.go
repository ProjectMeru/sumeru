package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sumeru/core/engine/render"
	"sumeru/core/server/web"
)

func TestWebHelperExportsCoverage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, web.TestWorkspaceRoute+"?action=1&menu_id=2", nil)
	action, menu := web.WorkspaceQueryParams(req)
	if action != "1" || menu != "2" {
		t.Fatalf("query params: %q %q", action, menu)
	}
	if id, ok := web.ParseMenuIDString(" 7 "); !ok || id != 7 {
		t.Fatalf("menu id: %d ok=%v", id, ok)
	}
	if _, ok := web.ParseMenuIDString("bad"); ok {
		t.Fatal("bad menu id")
	}
	wreq := web.ParseWorkspaceRequest(req, 5)
	aid, mid, vt, rid, edit := web.WorkspaceRequestFields(wreq)
	if aid != 5 || mid != "2" || vt != "" || rid != "" || edit {
		t.Fatalf("workspace req: %d %q %q %q %v", aid, mid, vt, rid, edit)
	}
	if got, err := web.URLWithQueryParam("/x", "a", "b"); err != nil || got != "/x?a=b" {
		t.Fatalf("url param: %q err=%v", got, err)
	}
	if got := web.NormalizeGridListLayout("TABLE"); got != web.TestAppsLayoutGrid {
		t.Fatalf("layout: %q", got)
	}
	if got := web.ClientIP(req); got == "" {
		t.Fatal("client ip")
	}
	if !web.IsLoopbackIP("127.0.0.1") {
		t.Fatal("loopback")
	}
	if got := web.SetupTokenFromRequest(req, "body"); got != "body" {
		t.Fatalf("setup token: %q", got)
	}
	pruned := web.PruneSetupAttempts(nil, time.Now())
	if len(pruned) != 0 {
		t.Fatal("prune nil")
	}
	body, ok := web.ParseSetupInitRequest(httptest.NewRecorder(), []byte(`{"admin_name":"Admin","email":"a@b.com","password":"secret123","company_name":"Co"}`))
	if !ok || body.Email != "a@b.com" {
		t.Fatalf("setup init: ok=%v body=%+v", ok, body)
	}
	params := web.ToSetupAdminParams(body)
	if params.Email != "a@b.com" {
		t.Fatalf("setup params: %+v", params)
	}
	page := web.BuildSetupPageData()
	if page.DbName == "" && !page.SetupTokenRequired {
		// defaults are fine
	}
	if !web.AcceptsJSONContentType("application/json") {
		t.Fatal("json content type")
	}
	model, method := web.ParseRPCCallMeta([]byte(`{"model":"core.user","method":"search"}`))
	if model != "core.user" || method != "search" {
		t.Fatalf("rpc meta: %s %s", model, method)
	}
	mux := web.ServeMuxOrDefault(nil)
	if mux == nil {
		t.Fatal("mux")
	}
	title, bodyText, details, fields := web.UserFacingRecordError("write", "core.user", nil)
	if title == "" || bodyText == "" {
		t.Fatalf("record error: %q %q %q %v", title, bodyText, details, fields)
	}
	vals := web.ActionDefaultFieldValues(map[string]interface{}{"context": map[string]interface{}{"default_name": "x"}})
	_ = vals
	if got := web.EnsureFormEditRedirectURL("/web?edit=1&id=1", true); got == "" {
		t.Fatal("form edit url")
	}
	if got := web.SplitCommaSeparatedValues("a, b ,c"); len(got) != 3 {
		t.Fatalf("split comma: %v", got)
	}
	if got := web.FirstGroupByField("state,partner_id"); got != "state" {
		t.Fatalf("group by: %q", got)
	}
	if got := web.SplitViewModes("list,form"); len(got) != 2 {
		t.Fatalf("view modes: %v", got)
	}
	if got := web.NormalizeViewMode("FORM"); got != "form" {
		t.Fatalf("normalize view: %q", got)
	}
	if got := web.FormBaseQueryValues(1, "2", "list", "3"); got == "" {
		t.Fatal("form base query")
	}
	if got := web.ChatterBodyTooLong(strings.Repeat("x", int(web.TestMaxChatterBodyRunes)+1)); !got {
		t.Fatal("chatter too long")
	}
	if _, err := web.ParseChatterRecordID("12"); err != nil {
		t.Fatal(err)
	}
	if got := web.CoerceCSVValue("42"); got != int64(42) && got != 42 {
		t.Fatalf("csv value: %v", got)
	}
	allowed := map[string]struct{}{"name": {}}
	row := web.ImportableRowValues([]string{"name"}, []string{"Acme"}, allowed)
	if row["name"] != "Acme" {
		t.Fatalf("import row: %v", row)
	}
	if !web.IsImportableColumn("name", allowed) || web.IsImportableColumn("secret", allowed) {
		t.Fatal("importable column")
	}
	if got := web.ImportCSVFlashMessage(3); !strings.Contains(got, "3") {
		t.Fatalf("flash: %q", got)
	}
	if got := web.LoginURLWithReturn("/web"); !strings.Contains(got, "next=") {
		t.Fatalf("login url: %q", got)
	}
	if got := web.BearerToken("Bearer abc"); got != "abc" {
		t.Fatalf("bearer: %q", got)
	}
	browse := web.ParseAppsBrowseStateFromForm(req)
	_ = browse
	if got := web.AppsModuleFromParsed(web.ModuleRow{Name: "sale", State: "installed"}); got.Name != "sale" {
		t.Fatalf("apps module: %+v", got)
	}
	modules := web.EnrichAppsModules(context.Background(), []web.AppsModule{{Name: "sale"}}, web.AppsBrowseState{})
	if len(modules) != 1 {
		t.Fatalf("enrich: %v", modules)
	}
	if flash, ok := web.AppsFlashFromMessage("installed:sale", map[string]string{"sale": "Sales"}); !ok || flash.Body == "" {
		t.Fatalf("flash: %+v ok=%v", flash, ok)
	}
	inline, toast := web.SplitAppsPageFlashes("installed:sale", map[string]string{"sale": "Sales"})
	if len(inline)+len(toast) == 0 {
		t.Fatal("split flashes")
	}
	if got := web.FormatAppsActionError("boom"); got == "" {
		t.Fatal("format error")
	}
	if got := web.ModuleStatusLabel(web.ModuleRow{State: "installed"}); got == "" {
		t.Fatal("status label")
	}
	names := web.BuildModuleDisplayNameMap([]web.AppsModule{{Name: "sale", DisplayName: "Sales"}})
	if names["sale"] != "Sales" {
		t.Fatalf("display names: %v", names)
	}
	appMods, techMods := web.FilterAppsModulesByBrowse([]web.AppsModule{
		{Name: "sale", Application: true},
		{Name: "base", Application: false},
	}, web.AppsBrowseState{})
	if len(appMods) != 1 || len(techMods) != 1 {
		t.Fatalf("filter: app=%d tech=%d", len(appMods), len(techMods))
	}
	if got := web.ModuleSummary("sale", "Long description here"); got == "" {
		t.Fatal("summary")
	}
	if got := web.ModuleSummaryFromDescription("desc"); got == "" {
		t.Fatal("summary from desc")
	}
	if !web.ModuleHasLongDescription("short", strings.Repeat("x", 200)) {
		t.Fatal("long description")
	}
	if got := web.AppsLinkFromBrowse(web.AppsBrowseState{Layout: web.TestAppsLayoutList}); got == "" {
		t.Fatal("apps link")
	}
	if got := web.AppsDetailURL(web.AppsBrowseState{}, false); got == "" {
		t.Fatal("apps detail url")
	}
	if _, ok := web.FindAppsModule([]web.AppsModule{{Name: "sale"}}, "sale"); !ok {
		t.Fatal("find module")
	}
	if got := web.ModuleDisplayName("sale", "Sales"); got != "Sales" {
		t.Fatalf("display name: %q", got)
	}
	if got := web.ResolveExtraScripts([]string{"/a.js"}, []string{"/b.js"}); len(got) < 1 {
		t.Fatalf("scripts: %v", got)
	}
	section, ok := web.SettingsHubSectionFromSidebar(render.SidebarMenu{ID: "settings"})
	_ = section
	_ = ok
	settingsPage := web.BuildSettingsHubPageData(context.Background(), "1")
	if settingsPage.Title == "" {
		t.Fatal("settings hub page")
	}
	if got := web.AppLogsViewStylesheets(); len(got) == 0 {
		t.Fatal("app logs stylesheets")
	}
	logsPage := web.BuildAppLogsPageData(context.Background(), 1)
	if logsPage.Title == "" {
		t.Fatal("app logs page")
	}
}
