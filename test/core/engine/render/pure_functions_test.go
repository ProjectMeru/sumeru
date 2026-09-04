package render_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)

func TestWorkspaceURL_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		q    render.WorkspaceQuery
		want string
	}{
		{"empty", render.WorkspaceQuery{}, "/web"},
		{"action id", render.WorkspaceQuery{ActionID: 5, MenuID: "2", ViewType: "list"},
			"/web?action=5&menu_id=2&view_type=list"},
		{"string action", render.WorkspaceQuery{Action: "sale.action", ViewType: "form"},
			"/web?action=sale.action&view_type=form"},
		{"record id", render.WorkspaceQuery{ActionID: 1, RecordID: "42", ViewType: "form"},
			"/web?action=1&id=42&view_type=form"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := render.WorkspaceURL(tc.q)
			if got != tc.want {
				t.Fatalf("WorkspaceURL = %q want %q", got, tc.want)
			}
			qs := render.WorkspaceQueryString(tc.q)
			if tc.want == "/web" && qs != "" {
				t.Fatalf("WorkspaceQueryString empty query: %q", qs)
			}
			if tc.want != "/web" && !strings.HasSuffix(tc.want, qs) {
				t.Fatalf("query string %q not in url %q", qs, tc.want)
			}
		})
	}
}

func TestSafeImageSrc_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src  string
		want bool
	}{
		{"https://cdn.example/logo.png", true},
		{"http://localhost/x", true},
		{"/static/img.png", true},
		{"data:image/png;base64,abc", true},
		{"", false},
		{"javascript:alert(1)", false},
		{"ftp://x", false},
	}
	for _, tc := range tests {
		if got := render.SafeImageSrc(tc.src); got != tc.want {
			t.Errorf("SafeImageSrc(%q) = %v want %v", tc.src, got, tc.want)
		}
	}
}

func TestSafeIframeURL_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src  string
		want bool
	}{
		{"https://reports.example/x", true},
		{"/web/report/1", true},
		{"http://reports.example/x", false},
		{"javascript:alert(1)", false},
		{"data:text/html,hi", false},
		{"//evil.example", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := render.SafeIframeURL(tc.src); got != tc.want {
			t.Errorf("SafeIframeURL(%q) = %v want %v", tc.src, got, tc.want)
		}
	}
	if !render.SafeIframeURLAllowHTTP("http://reports.example/x") {
		t.Fatal("SafeIframeURLAllowHTTP should allow http")
	}
}

func TestFieldDisplayLabel_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		field parser.Field
		want  string
	}{
		{parser.Field{Name: "partner_id", Label: "Customer"}, "Customer"},
		{parser.Field{Name: "date_order"}, "Date Order"},
		{parser.Field{Name: "x"}, "X"},
	}
	for _, tc := range tests {
		if got := render.FieldDisplayLabel(tc.field); got != tc.want {
			t.Errorf("FieldDisplayLabel(%+v) = %q want %q", tc.field, got, tc.want)
		}
	}
}

func TestIconHueFromString_stable(t *testing.T) {
	t.Parallel()
	a := render.IconHueFromString("sale")
	b := render.IconHueFromString("sale")
	if a != b || a < 0 || a >= 360 {
		t.Fatalf("IconHueFromString sale: %d %d", a, b)
	}
	if got := render.IconHueFromString(""); got != 265 {
		t.Fatalf("empty name default hue: %d", got)
	}
}

func TestIconLetterFromName(t *testing.T) {
	t.Parallel()
	if got := render.IconLetterFromName("Sales"); got != "S" {
		t.Fatalf("IconLetterFromName: %q", got)
	}
	if got := render.IconLetterFromName("  "); got != "?" {
		t.Fatalf("blank: %q", got)
	}
}

func TestModuleIconServePath_emptyWithoutAddon(t *testing.T) {
	t.Parallel()
	if got := render.ModuleIconServePath("nonexistent_module_xyz", "static/icon.png"); got != "" {
		t.Fatalf("missing addon should return empty: %q", got)
	}
	if got := render.ModuleIconServePath("nonexistent_module_xyz", "/etc/passwd"); got != "" {
		t.Fatalf("absolute path must be rejected, got %q", got)
	}
	if got := render.ModuleIconServePath("nonexistent_module_xyz", "../../../etc/passwd"); got != "" {
		t.Fatalf("traversal path must be rejected, got %q", got)
	}
}

func TestAppsViewTabs_andHomeViewTabs(t *testing.T) {
	t.Parallel()
	tabs := render.AppsViewTabs("grid", "installed", "sale", "installed", "apps", "crm", "sales", "category")
	if len(tabs) != 2 || !tabs[0].Active {
		t.Fatalf("AppsViewTabs: %+v", tabs)
	}
	homeTabs := render.HomeViewTabs("list")
	if len(homeTabs) == 0 {
		t.Fatal("HomeViewTabs empty")
	}
}

func TestBreadcrumbBuilders_noDB(t *testing.T) {
	ctx := context.Background()
	apps := render.BuildAppsBreadcrumbs(ctx, "/web/apps", "Sales")
	if len(apps) != 3 || apps[2].Label != "Sales" {
		t.Fatalf("BuildAppsBreadcrumbs detail: %+v", apps)
	}
	appsList := render.BuildAppsBreadcrumbs(ctx, "", "")
	if len(appsList) != 2 || appsList[1].Href != "" {
		t.Fatalf("BuildAppsBreadcrumbs list: %+v", appsList)
	}
	dash := render.BuildHomeDashboardBreadcrumbs(ctx)
	if len(dash) != 2 || dash[1].Label != "Dashboard" {
		t.Fatalf("BuildHomeDashboardBreadcrumbs: %+v", dash)
	}
	settings := render.BuildSettingsHubBreadcrumbs(ctx)
	if len(settings) != 1 || settings[0].Label != "Settings" {
		t.Fatalf("BuildSettingsHubBreadcrumbs: %+v", settings)
	}
	if render.SettingsHomeURL() != "/web/settings" {
		t.Fatal("SettingsHomeURL")
	}
	if got := render.MenuWebURL(3, 10); !strings.Contains(got, "menu_id=3") || !strings.Contains(got, "action=10") {
		t.Fatalf("MenuWebURL: %q", got)
	}
}

func TestNormalizeImageCrop_clamps(t *testing.T) {
	t.Parallel()
	c := render.NormalizeImageCrop(render.ImageCrop{X: -5, Y: 200, Zoom: 10})
	if c.X != 0 || c.Y != 100 || c.Zoom != 4 {
		t.Fatalf("NormalizeImageCrop: %+v", c)
	}
	if attr := render.AvatarCropStyle(c, false); string(attr) != "" {
		t.Fatalf("inactive AvatarCropStyle should be empty: %q", attr)
	}
}

func TestSplitModuleMenuChain_empty(t *testing.T) {
	t.Parallel()
	_, _, ok := render.SplitModuleMenuChainForTest(nil)
	if ok {
		t.Fatal("empty chain should not ok")
	}
}

func TestWorkspaceRecordBreadcrumbLabel_formRecord(t *testing.T) {
	t.Parallel()
	label := render.WorkspaceRecordBreadcrumbLabelForTest(render.BreadcrumbInput{
		ViewType: render.ViewModeForm,
		RecordID: 1,
		Record:   map[string]interface{}{"name": "Acme"},
	})
	if label != "Acme" {
		t.Fatalf("record label: %q", label)
	}
	label = render.WorkspaceRecordBreadcrumbLabelForTest(render.BreadcrumbInput{
		ViewType: render.ViewModeForm,
		RecordID: 1,
	})
	if label != "Record" {
		t.Fatalf("fallback record label: %q", label)
	}
}

func TestBuildSidebarMenus_deniesAll(t *testing.T) {
	t.Parallel()
	menus := []parser.MenuItem{
		{ID: "1", Name: "Root"},
		{ID: "2", Name: "Child", ParentID: "1", Action: "/web"},
	}
	out := render.BuildSidebarMenus(menus, "1", func(parser.MenuItem) bool { return false })
	if len(out) != 0 {
		t.Fatalf("deny all: %+v", out)
	}
}

func TestResolveActiveModuleID_walksUp(t *testing.T) {
	t.Parallel()
	menus := []parser.MenuItem{
		{ID: "1", Name: "App"},
		{ID: "2", Name: "Sub", ParentID: "1"},
		{ID: "3", Name: "Leaf", ParentID: "2"},
	}
	if got := render.ResolveActiveModuleID(menus, "3"); got != "1" {
		t.Fatalf("module root: %q", got)
	}
	if got := render.ResolveActiveModuleID(menus, "missing"); got != "" {
		t.Fatalf("missing menu: %q", got)
	}
}
