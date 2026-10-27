package sdk_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"sumeru/core/orm"
	"sumeru/core/sdk"
)

func TestEnsureOne(t *testing.T) {
	if _, err := sdk.Browse(context.Background(), "demo").EnsureOne(); err == nil {
		t.Fatal("empty set")
	}
	if _, err := sdk.Browse(context.Background(), "demo", 1, 2).EnsureOne(); err == nil {
		t.Fatal("two ids")
	}
	id, err := sdk.Browse(context.Background(), "demo", 4).EnsureOne()
	if err != nil || id != 4 {
		t.Fatalf("id=%d err=%v", id, err)
	}
}

func TestSearchReadSelectsRequestedColumns(t *testing.T) {
	mock := setupSDKMock(t, "test.sdk.read")
	mock.ExpectQuery(`SELECT "name" FROM "test_sdk_read" WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Alpha"))

	rows, err := sdk.SearchRead(orm.ContextWithBypass(context.Background(), true), "test.sdk.read", nil, []string{"name"}, 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0]["name"] != "Alpha" {
		t.Fatalf("rows %+v", rows)
	}
	if _, ok := rows[0]["active"]; ok {
		t.Fatal("unexpected column")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSearchReadRejectsUnknownField(t *testing.T) {
	setupSDKMock(t, "test.sdk.read.bad")
	_, err := sdk.SearchRead(context.Background(), "test.sdk.read.bad", nil, []string{"missing"}, 0, 0, "")
	if err == nil {
		t.Fatal("expected unknown field")
	}
}

func TestWriteManyIDsOneUpdate(t *testing.T) {
	const model = "test.sdk.write"
	mock := setupSDKMock(t, model)
	ctx := orm.ContextWithBypass(context.Background(), true)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "test_sdk_write" WHERE .*"id" IN \(\$1,\$2\).*FOR UPDATE`).
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "A").AddRow(2, "B"))
	mock.ExpectExec(`UPDATE "test_sdk_write"`).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	n, err := sdk.Write(ctx, model, []int{1, 2}, map[string]interface{}{"name": "N"})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("count %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSearchWriteReturnsCount(t *testing.T) {
	const model = "test.sdk.searchwrite"
	mock := setupSDKMock(t, model)
	ctx := orm.ContextWithBypass(context.Background(), true)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "test_sdk_searchwrite" WHERE .*"active" = \$1.*FOR UPDATE`).
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "active"}).AddRow(3, "A", true))
	mock.ExpectExec(`UPDATE "test_sdk_searchwrite"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	n, err := sdk.SearchWrite(ctx, model, [][]interface{}{{"active", "=", true}}, map[string]interface{}{"name": "N"})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("count %d", n)
	}
}

func TestRegisterMethodOneID(t *testing.T) {
	const model = "test.sdk.method"
	mock := setupSDKMock(t, model)
	mock.ExpectQuery(`SELECT \* FROM "test_sdk_method"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(7, "A"))

	var got []int
	sdk.RegisterMethod(model, "do_it", func(_ context.Context, rs sdk.Recordset) (sdk.ActionResult, error) {
		got = rs.IDs()
		return sdk.ActionResult{}, nil
	})
	if _, err := orm.RunObjectAction(orm.ContextWithBypass(context.Background(), true), model, 7, "do_it", nil); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != 7 {
		t.Fatalf("ids %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultGetOverlaysFieldDefaults(t *testing.T) {
	const model = "test.sdk.defaults"
	orm.RegisterStubModelForTest(t, model, []orm.FieldDefinition{
		{Name: "name", Type: orm.Char, DefaultVal: "draft"},
		{Name: "owner_id", Type: orm.Integer, DefaultVal: "current_user"},
		{Name: "active", Type: orm.Boolean, DefaultVal: true},
	})
	ctx := orm.ContextWithUID(context.Background(), 1)
	got, err := sdk.DefaultGet(ctx, model, nil, map[string]interface{}{"name": "from-action", "ignored": true})
	if err != nil {
		t.Fatal(err)
	}
	if got["name"] != "from-action" || got["active"] != true || got["owner_id"] != 1 {
		t.Fatalf("defaults %+v", got)
	}
	if _, ok := got["ignored"]; ok {
		t.Fatal("overlay kept an unknown field")
	}
}

func TestCopyUsesCreate(t *testing.T) {
	const model = "test.sdk.copy"
	mock := setupSDKMock(t, model)
	ctx := orm.ContextWithBypass(context.Background(), true)
	mock.ExpectQuery(`SELECT \* FROM "test_sdk_copy"`).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code"}).AddRow(5, "Alpha", "A"))
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "test_sdk_copy"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(6))
	mock.ExpectCommit()

	id, err := sdk.Copy(ctx, model, 5)
	if err != nil {
		t.Fatal(err)
	}
	if id != 6 {
		t.Fatalf("copy id %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func setupSDKMock(t *testing.T, model string) sqlmock.Sqlmock {
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
	orm.RegisterStubModelForTest(t, model, []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "code", Type: orm.Char, Unique: true},
		{Name: "active", Type: orm.Boolean},
	})
	return mock
}
