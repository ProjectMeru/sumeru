package orm_test

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"sumeru/core/orm"
)

func TestReadGroupDirectDeniesGroupedField(t *testing.T) {
	const model = "test.readgroup.acl"
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	orm.SetDBForTest(orm.NewDBWrapper(db))
	t.Cleanup(func() {
		_ = db.Close()
		orm.ResetDBForTest()
		delete(orm.Registry, model)
	})
	orm.RegisterModelWithModule(safetyModel{
		name: model,
		fields: []orm.FieldDefinition{
			{Name: "name", Type: orm.Char},
			{Name: "secret", Type: orm.Char, Groups: "base.group_system"},
		},
	}, "test")

	uid := 424242
	orm.InvalidateEffectiveGroups(context.Background(), uid)
	emptyGroups := sqlmock.NewRows([]string{"group_id"})
	mock.ExpectQuery(`SELECT group_id FROM "core_group_user_rel"`).
		WithArgs(uid).
		WillReturnRows(emptyGroups)
	mock.ExpectQuery(`SELECT group_id, "perm_read" FROM "sys_access"`).
		WithArgs(model).
		WillReturnRows(sqlmock.NewRows([]string{"group_id", "perm_read"}).AddRow(nil, true))
	mock.ExpectQuery(`SELECT group_id FROM "core_group_user_rel"`).
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"group_id"}))
	mock.ExpectQuery(`SELECT field_name, group_id, "perm_read" FROM "sys_field_access"`).
		WithArgs(model).
		WillReturnRows(sqlmock.NewRows([]string{"field_name", "group_id", "perm_read"}))
	mock.ExpectQuery(`SELECT group_id FROM "core_group_user_rel"`).
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"group_id"}))

	_, err = orm.ReadGroupDirect(orm.ContextWithUID(context.Background(), uid), model, nil, orm.ReadGroupSpec{
		GroupBy: []string{"secret"},
		Fields:  []orm.ReadGroupField{{Name: "secret_sum", Field: "secret", Measure: "sum"}},
	})
	if err == nil || !strings.Contains(err.Error(), "field access denied") {
		t.Fatalf("err %v", err)
	}
}
