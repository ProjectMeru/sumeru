package web_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"sumeru/core/orm"
	"sumeru/core/server/web"
)

func setupURLActionWebTest(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	orm.SetDBForTest(orm.NewDBWrapper(db))
	t.Cleanup(func() {
		_ = db.Close()
		orm.ResetDBForTest()
	})
	orm.RegisterStubModelForTest(t, "sys.model.data", []orm.FieldDefinition{
		{Name: "module", Type: orm.Char},
		{Name: "name", Type: orm.Char},
		{Name: "model", Type: orm.Char},
		{Name: "core_id", Type: orm.Integer},
	})
	orm.RegisterStubModelForTest(t, "sys.action.url", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "url", Type: orm.Char},
	})
	return mock
}

func bypassCtx() context.Context {
	return orm.ContextWithBypass(context.Background(), true)
}

func TestResolveNavigationActionURLRedirect(t *testing.T) {
	mock := setupURLActionWebTest(t)
	metaRows := sqlmock.NewRows([]string{"id", "module", "name", "model", "core_id"}).
		AddRow(1, "account", "x", "sys.action.url", 5)
	mock.ExpectQuery(`SELECT \* FROM "sys_model_data" WHERE \("core_id" = \$1 AND "model" IN \(\$2,\$3\)\)`).
		WithArgs(5, "sys.action.window", "sys.action.url").
		WillReturnRows(metaRows)
	urlRows := sqlmock.NewRows([]string{"id", "name", "url"}).
		AddRow(5, "Profit & Loss", "/account/reports/view?type=profit_loss")
	mock.ExpectQuery(`SELECT \* FROM "sys_action_url" WHERE \("id" = \$1\) LIMIT 1`).
		WithArgs(5).
		WillReturnRows(urlRows)

	kind, url, err := web.ResolveNavigationActionForTest(bypassCtx(), 5, "")
	if err != nil {
		t.Fatal(err)
	}
	if kind != int(web.NavActionURLForTest) {
		t.Fatalf("kind=%d want url", kind)
	}
	if url != "/account/reports/view?type=profit_loss" {
		t.Fatalf("url=%q", url)
	}
}

func TestResolveNavigationActionWindowUnchanged(t *testing.T) {
	mock := setupURLActionWebTest(t)
	orm.RegisterStubModelForTest(t, "sys.action.window", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "core_model", Type: orm.Char},
		{Name: "view_mode", Type: orm.Char},
	})
	metaRows := sqlmock.NewRows([]string{"id", "module", "name", "model", "core_id"}).
		AddRow(1, "base", "x", "sys.action.window", 3)
	mock.ExpectQuery(`SELECT \* FROM "sys_model_data" WHERE \("core_id" = \$1 AND "model" IN \(\$2,\$3\)\)`).
		WithArgs(3, "sys.action.window", "sys.action.url").
		WillReturnRows(metaRows)
	windowRows := sqlmock.NewRows([]string{"id", "name", "core_model", "view_mode"}).
		AddRow(3, "Contacts", "core.partner", "list,form")
	mock.ExpectQuery(`SELECT \* FROM "sys_action_window" WHERE \("id" = \$1\) LIMIT 1`).
		WithArgs(3).
		WillReturnRows(windowRows)

	kind, url, err := web.ResolveNavigationActionForTest(bypassCtx(), 3, "")
	if err != nil {
		t.Fatal(err)
	}
	if kind != int(web.NavActionWindowForTest) || url != "" {
		t.Fatalf("kind=%d url=%q", kind, url)
	}
}

func TestBuildIframeSwcPayload(t *testing.T) {
	got := web.BuildIframeSwcPayloadForTest(bypassCtx(), 9, "12", "/account/reports/view?type=profit_loss")
	if got["viewType"] != "iframe" {
		t.Fatalf("viewType=%v", got["viewType"])
	}
	if got["iframeUrl"] != "/account/reports/view?type=profit_loss" {
		t.Fatalf("iframeUrl=%v", got["iframeUrl"])
	}
	if got["model"] != "sys.action.url" {
		t.Fatalf("model=%v", got["model"])
	}
}
