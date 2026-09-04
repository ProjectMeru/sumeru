package orm_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"sumeru/core/orm"
)

func TestResolveActionRecord_numericURLAction(t *testing.T) {
	mock := setupMockORM(t)
	orm.RegisterStubModelForTest(t, "sys.model.data", []orm.FieldDefinition{
		{Name: "module", Type: orm.Char},
		{Name: "name", Type: orm.Char},
		{Name: "model", Type: orm.Char},
		{Name: "core_id", Type: orm.Integer},
	})
	rows := sqlmock.NewRows([]string{"id", "module", "name", "model", "core_id"}).
		AddRow(1, "account", "action_report_pl", "sys.action.url", 42)
	mock.ExpectQuery(`SELECT \* FROM "sys_model_data" WHERE \("core_id" = \$1 AND "model" IN \(\$2,\$3\)\)`).
		WithArgs(42, "sys.action.window", "sys.action.url").
		WillReturnRows(rows)

	modelName, coreID, err := orm.ResolveActionRecordForTest(bypassCtx(), 42, "")
	if err != nil {
		t.Fatal(err)
	}
	if modelName != "sys.action.url" || coreID != 42 {
		t.Fatalf("got model=%q id=%d", modelName, coreID)
	}
}

func TestResolveActionRecord_xmlID(t *testing.T) {
	mock := setupMockORM(t)
	orm.RegisterStubModelForTest(t, "sys.model.data", []orm.FieldDefinition{
		{Name: "module", Type: orm.Char},
		{Name: "name", Type: orm.Char},
		{Name: "model", Type: orm.Char},
		{Name: "core_id", Type: orm.Integer},
	})
	rows := sqlmock.NewRows([]string{"id", "module", "name", "model", "core_id"}).
		AddRow(1, "account", "action_report_pl", "sys.action.url", 7)
	mock.ExpectQuery(`SELECT \* FROM "sys_model_data" WHERE \("module" = \$1 AND "name" = \$2\) LIMIT 1`).
		WithArgs("account", "action_report_pl").
		WillReturnRows(rows)

	modelName, coreID, err := orm.ResolveActionRecordForTest(bypassCtx(), 0, "account.action_report_pl")
	if err != nil {
		t.Fatal(err)
	}
	if modelName != "sys.action.url" || coreID != 7 {
		t.Fatalf("got model=%q id=%d", modelName, coreID)
	}
}
