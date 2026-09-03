package coverage_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"sumeru/addons/automation"
	"sumeru/core/applog"
	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
	"sumeru/core/engine/swcmeta"
	"sumeru/core/event"
	"sumeru/core/importgen"
	"sumeru/core/modelmeta"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/report"
	"sumeru/core/scheduler"
	"sumeru/core/server"
	"sumeru/core/server/api"
	"sumeru/core/server/router"
	"sumeru/core/server/web"
	"sumeru/test/harness"
)

func TestForTestExports_applog(t *testing.T) {
	t.Parallel()
	loc, name := applog.ParseLogTimezone("UTC")
	if loc == nil || name == "" {
		t.Fatalf("ParseLogTimezone UTC: loc=%v name=%q", loc, name)
	}
	applog.SetLogLocationForTest(loc)
	if got := applog.EffectiveLocationForTest(); got != loc {
		t.Fatalf("EffectiveLocationForTest = %v", got)
	}
}

func TestForTestExports_importgen(t *testing.T) {
	root := harness.RepoRoot(t)
	body := importgen.RenderZRefsForTest([]importgen.ExportedRefForTest{
		{Name: "TestModel", Kind: "phantom", TechnicalModel: "test.model"},
	})
	if body == "" {
		t.Fatal("RenderZRefsForTest empty")
	}
	models, err := importgen.ParseDirModelsForTest(filepath.Join(root, "test/core/modelreg"))
	if err != nil {
		t.Fatal(err)
	}
	_ = models
	names, err := importgen.ScanPackageForModelsForTest(filepath.Join(root, "core", "modelreg", "testdata", "selection"))
	if err != nil {
		t.Fatal(err)
	}
	_ = names
}

func TestForTestExports_modelmeta(t *testing.T) {
	t.Parallel()
	sel, def := modelmeta.ExtractDefaultFromSelectionTailForTest("a:A,b:B,default=a")
	if def != "a" || sel == "" {
		t.Fatalf("ExtractDefaultFromSelectionTailForTest: sel=%q def=%q", sel, def)
	}
	body, selection := modelmeta.PeelSelectionTagForTest("required,selection=x:X")
	if body == "" || selection == "" {
		t.Fatalf("PeelSelectionTagForTest: body=%q selection=%q", body, selection)
	}
}

func TestForTestExports_modelreg(t *testing.T) {
	t.Parallel()
	if got := modelreg.ParseSelectionForTest("low:Low,high:High"); len(got) != 2 {
		t.Fatalf("ParseSelectionForTest: %v", got)
	}
	if modelreg.ParseDefaultForTest("42", orm.Integer) != int64(42) {
		t.Fatal("ParseDefaultForTest integer")
	}
	ft, widget, err := modelreg.MapMarkerTypeForTest("String")
	if err != nil || ft != orm.Char || widget != "" {
		t.Fatalf("MapMarkerTypeForTest String: ft=%s widget=%q err=%v", ft, widget, err)
	}
	ctx := modelreg.NewRegisterCtxForTest()
	type partner struct{}
	ctx.SetTypeMapping("Many2One", "res.partner", "Partner", reflect.TypeOf(partner{}))
}

func TestForTestExports_render(t *testing.T) {
	t.Parallel()
	chain := []render.MenuCrumbForTest{
		{ID: 1, Name: "CRM"},
		{ID: 2, Name: "Leads"},
	}
	mod, menus, ok := render.SplitModuleMenuChainForTest(chain)
	if !ok || mod.Name != "CRM" || len(menus) != 1 {
		t.Fatalf("SplitModuleMenuChainForTest: mod=%+v menus=%v ok=%v", mod, menus, ok)
	}
	if got := render.WorkspaceViewBreadcrumbLabelForTest(render.ViewModeGraph); got != "Graph" {
		t.Fatalf("WorkspaceViewBreadcrumbLabelForTest graph: %q", got)
	}
	label := render.WorkspaceRecordBreadcrumbLabelForTest(render.BreadcrumbInput{
		ViewType: render.ViewModeForm,
		ResModel: "core.user",
		RecordID: 0,
	})
	if label == "" {
		t.Fatal("WorkspaceRecordBreadcrumbLabelForTest empty")
	}
	set := render.ViewModeFilterSetForTest([]string{"list", "form"})
	if _, ok := set["list"]; !ok {
		t.Fatalf("ViewModeFilterSetForTest: %v", set)
	}
}

func TestForTestExports_scheduler(t *testing.T) {
	scheduler.ClearCronHandlers()
	t.Cleanup(scheduler.ClearCronHandlers)
	var called bool
	scheduler.RegisterCronHandler("export_job", func(_ context.Context, _ map[string]interface{}) error {
		called = true
		return nil
	})
	scheduler.ExecuteCronForTest(context.Background(), scheduler.CronRunInput{
		ID: 9, Name: "Export", Code: "export_job", EventName: "export.tick",
	})
	if !called {
		t.Fatal("ExecuteCronForTest did not invoke handler")
	}
}

func TestForTestExports_server(t *testing.T) {
	t.Parallel()
	if got := server.ListenAddrForTest("127.0.0.1", "9090"); got != "127.0.0.1:9090" {
		t.Fatalf("ListenAddrForTest: %q", got)
	}
	cfg := server.ConfigForTest("", "8080", true)
	if got := server.SetupListenAddrForTest(cfg); got != "127.0.0.1:8080" {
		t.Fatalf("SetupListenAddrForTest: %q", got)
	}
	rel, ok := server.NormalizeManifestAssetRelForTest("static/app.js")
	if !ok || rel != "static/app.js" {
		t.Fatalf("NormalizeManifestAssetRelForTest: %q ok=%v", rel, ok)
	}
	url := server.ManifestAssetPublicURLForTest("base", "static/app.js")
	if url == "" {
		t.Fatal("ManifestAssetPublicURLForTest empty")
	}
}

func TestForTestExports_router(t *testing.T) {
	router.Clear()
	t.Cleanup(router.Clear)
	router.UpsertRouteForTest(router.Route{
		Method: http.MethodGet,
		Path:   "/export-test",
		Handler: func(http.ResponseWriter, *http.Request) {},
	})
	if router.RegisteredRouteCountForTest() != 1 {
		t.Fatalf("RegisteredRouteCountForTest = %d", router.RegisteredRouteCountForTest())
	}
}

func TestForTestExports_report(t *testing.T) {
	t.Parallel()
	if got := report.SanitizeSpreadsheetCellForTest("=SUM(A1)"); got[0] != '\t' {
		t.Fatalf("SanitizeSpreadsheetCellForTest: %q", got)
	}
	data, err := report.ExportXLSXForTest([]string{"A"}, [][]string{{"1"}})
	if err != nil || len(data) == 0 {
		t.Fatalf("ExportXLSXForTest: len=%d err=%v", len(data), err)
	}
	csv, err := report.WriteCSVForTest([]string{"A"}, [][]string{{"x"}})
	if err != nil || len(csv) == 0 {
		t.Fatalf("WriteCSVForTest: err=%v", err)
	}
	sheet := report.SheetXMLForTest([]string{"Col"}, [][]string{{"v"}})
	if sheet == "" {
		t.Fatal("SheetXMLForTest empty")
	}
}

func TestForTestExports_swcmeta(t *testing.T) {
	ctx := context.Background()
	sheet := swcmeta.SerializeSheetForTest(ctx, "test.model", &parser.Sheet{
		Field: []parser.Field{{Name: "name"}},
	})
	if sheet == nil || len(sheet.Fields) != 1 {
		t.Fatalf("SerializeSheetForTest: %+v", sheet)
	}
	group := swcmeta.SerializeGroupForTest(ctx, "test.model", parser.Group{Title: "G"})
	if group.String != "G" {
		t.Fatalf("SerializeGroupForTest: %+v", group)
	}
	div := swcmeta.SerializeDivForTest(ctx, "test.model", parser.Div{Class: "x"})
	if div.Class != "x" {
		t.Fatalf("SerializeDivForTest: %+v", div)
	}
	_ = swcmeta.FormMetaForModelForTest("missing.model")
	fields := swcmeta.SerializeFieldsForTest(ctx, []parser.Field{{Name: "x"}})
	if len(fields) != 1 {
		t.Fatal("SerializeFieldsForTest")
	}
	enriched := swcmeta.EnrichFieldForTest("test.model", swcmeta.ArchField{Name: "x"})
	if enriched.Name != "x" {
		t.Fatalf("EnrichFieldForTest: %+v", enriched)
	}
	if swcmeta.WorkspacePayloadTypeForTest().Name() != "WorkspacePayload" {
		t.Fatal("WorkspacePayloadTypeForTest")
	}
}

func TestForTestExports_api(t *testing.T) {
	t.Parallel()
	if api.StatusForCodeForTest(api.CodeUnauthorized) != http.StatusUnauthorized {
		t.Fatal("StatusForCodeForTest")
	}
	err := api.NewRPCErrorForTest(api.CodeInvalidArgs, "bad", nil)
	if code, ok := api.RPCErrorCodeForTest(err); !ok || code != api.CodeInvalidArgs {
		t.Fatalf("RPCErrorCodeForTest: code=%q ok=%v", code, ok)
	}
	if got := string(api.NormArgsForTest(nil)); got != "[]" {
		t.Fatalf("NormArgsForTest: %q", got)
	}
	arr, err := api.ParseArgsArrayForTest(json.RawMessage(`[1]`))
	if err != nil || len(arr) != 1 {
		t.Fatalf("ParseArgsArrayForTest: %v err=%v", arr, err)
	}
	d, err := api.ParseDomainArgForTest(json.RawMessage(`[["id","=",1]]`))
	if err != nil || len(d) != 1 {
		t.Fatalf("ParseDomainArgForTest: %v err=%v", d, err)
	}
	limit, offset := api.ParseLimitOffsetForTest(json.RawMessage(`{"limit":5,"offset":2}`))
	if limit != 5 || offset != 2 {
		t.Fatalf("ParseLimitOffsetForTest: limit=%d offset=%d", limit, offset)
	}
	if v, ok := api.ToFloatForTest(3.5); !ok || v != 3.5 {
		t.Fatalf("ToFloatForTest: v=%v ok=%v", v, ok)
	}
	rows := api.ProjectFieldsForTest([]map[string]interface{}{{"id": 1, "name": "a"}}, []string{"name"})
	if len(rows) != 1 || rows[0]["name"] != "a" {
		t.Fatalf("ProjectFieldsForTest: %+v", rows)
	}
	if err := api.ValidateKwargsForTest(json.RawMessage(`{"limit":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := api.CapRPCIDsForTest(make([]int, 10001)); err == nil {
		t.Fatal("CapRPCIDsForTest expected error")
	}
}

func TestForTestExports_orm(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.export.model", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "active", Type: orm.Boolean},
	})
	where, args, err := orm.BuildSearchWhereClauseForTest("test.export.model", [][]interface{}{
		{"name", "=", "x"},
	})
	if err != nil || where == "" {
		t.Fatalf("BuildSearchWhereClauseForTest: where=%q err=%v", where, err)
	}
	_ = args
	v, err := orm.CoerceFieldValueForTest(orm.FieldDefinition{Name: "active", Type: orm.Boolean}, "true")
	if err != nil || v != true {
		t.Fatalf("CoerceFieldValueForTest: v=%v err=%v", v, err)
	}
	orCount, leaves, ok := orm.SplitDomainORPrefixForTest([][]interface{}{
		{"|"}, {"name", "=", "a"}, {"name", "=", "b"},
	})
	if !ok || orCount != 1 || len(leaves) != 2 {
		t.Fatalf("SplitDomainORPrefixForTest: or=%d leaves=%d ok=%v", orCount, len(leaves), ok)
	}
	where, args, err = orm.BuildAndWhereClausesForTest("test.export.model", [][][]interface{}{
		{{"active", "=", true}},
	})
	if err != nil || where == "" {
		t.Fatalf("BuildAndWhereClausesForTest: err=%v", err)
	}
	_ = args
	if !orm.RecordMatchesDomainForTest("", map[string]interface{}{"name": "a"}, [][]interface{}{{"name", "=", "a"}}) {
		t.Fatal("RecordMatchesDomainForTest")
	}
	stub := orm.NewStubModelForTest("test.stub.only", nil)
	if stub.ModelName() != "test.stub.only" {
		t.Fatal("NewStubModelForTest")
	}
}

func TestForTestExports_automation(t *testing.T) {
	ctx := context.Background()
	event.Clear()
	t.Cleanup(event.Clear)
	var published string
	event.Subscribe("export.ping", func(_ context.Context, ev event.Event) error {
		published = ev.Name
		return nil
	})
	row := map[string]interface{}{"code": "publish:export.ping"}
	ev := event.Event{Name: "record.created", Payload: map[string]interface{}{"model": "crm.lead", "id": 1}}
	if err := automation.ExecuteServerActionForTest(ctx, row, ev); err != nil {
		t.Fatal(err)
	}
	if published != "export.ping" {
		t.Fatalf("ExecuteServerActionForTest publish: got %q", published)
	}
	if err := automation.RunServerActionsForEventForTest(ctx, ev); err != nil {
		t.Fatal(err)
	}
}

func TestForTestExports_web(t *testing.T) {
	web.ResetSetupRateLimiterForTest()
	req, ok := web.ParseSetupInitRequest(httptest.NewRecorder(), []byte(`{"admin_name":"Admin","email":"a@b.c","password":"secret","company_name":"Co"}`))
	if !ok || req.AdminName == "" {
		t.Fatalf("ParseSetupInitRequest: ok=%v req=%+v", ok, req)
	}
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(42)
	if got := web.HomeRouteWithMenuForTest("5"); got == "" {
		t.Fatal("HomeRouteWithMenuForTest")
	}
	modes := web.PrependViewModeForTest("form", []string{"list"})
	if len(modes) != 2 || modes[0] != "form" {
		t.Fatalf("PrependViewModeForTest: %v", modes)
	}
	if !web.IsNumericRecordIDForTest("99") {
		t.Fatal("IsNumericRecordIDForTest")
	}
	r := httptest.NewRequest(http.MethodGet, "/web?view_mode=list", nil)
	candidates := web.WorkspaceViewModeCandidatesForTest(r, map[string]interface{}{"view_mode": "list,form"})
	if len(candidates) == 0 {
		t.Fatal("WorkspaceViewModeCandidatesForTest")
	}
	tabs := web.ActionViewModesForTabsForTest(map[string]interface{}{"view_mode": "list,kanban"})
	if len(tabs) == 0 {
		t.Fatal("ActionViewModesForTabsForTest")
	}
	if id, ok := web.ParsePositiveRecordIDForTest("7"); !ok || id != 7 {
		t.Fatalf("ParsePositiveRecordIDForTest: id=%d ok=%v", id, ok)
	}
	sections := web.PartitionListSectionsForTest([]map[string]interface{}{
		{"state": "draft"}, {"state": "done"},
	}, "state")
	if len(sections) != 2 {
		t.Fatalf("PartitionListSectionsForTest: %d sections", len(sections))
	}
	hub := web.NewBusHubForTest()
	client := web.NewSwcBusClientForTest(1, 2)
	hub.Register(client)
	hub.Broadcast(1, []byte("ping"))
	select {
	case msg := <-client.Recv():
		if string(msg) != "ping" {
			t.Fatalf("bus broadcast: %q", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("bus broadcast timeout")
	}
	w := httptest.NewRecorder()
	web.SetRecordErrorFlashForTest(w, web.PageFlash{Title: "err"})
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range w.Result().Cookies() {
		r2.AddCookie(c)
	}
	flash, ok := web.ConsumeRecordErrorFlashForTest(r2, httptest.NewRecorder())
	if !ok || flash.Title != "err" {
		t.Fatalf("record error flash: ok=%v flash=%+v", ok, flash)
	}
}
