package orm_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"sumeru/core/orm"
)

func setupMockORM(t *testing.T) sqlmock.Sqlmock {
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
	orm.RegisterStubModelForTest(t, "test.mock", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "active", Type: orm.Boolean},
	})
	return mock
}

func bypassCtx() context.Context {
	return orm.ContextWithBypass(context.Background(), true)
}

func TestSearchWithMockDB(t *testing.T) {
	mock := setupMockORM(t)
	rows := sqlmock.NewRows([]string{"id", "name", "active"}).
		AddRow(1, "Alpha", true).
		AddRow(2, "Beta", false)
	mock.ExpectQuery(`SELECT \* FROM "test_mock" WHERE 1=1`).
		WillReturnRows(rows)

	got, err := orm.Search(bypassCtx(), "test.mock", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0]["name"] != "Alpha" {
		t.Fatalf("search: %+v", got)
	}
}

func TestSearchLimitWithMockDB(t *testing.T) {
	mock := setupMockORM(t)
	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(3, "Gamma")
	mock.ExpectQuery(`SELECT \* FROM "test_mock" WHERE 1=1 ORDER BY "id" ASC LIMIT \$1 OFFSET \$2`).
		WithArgs(5, 0).
		WillReturnRows(rows)

	got, err := orm.SearchLimit(bypassCtx(), "test.mock", nil, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("search limit: %+v", got)
	}
}

func TestUpdateRecordByIDWithMockDB(t *testing.T) {
	mock := setupMockORM(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "test_mock" WHERE \("id" = \$1\) FOR UPDATE`).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "Old"))
	mock.ExpectExec(`UPDATE "test_mock" SET "name" = \$1 WHERE \("id" = \$2\)`).
		WithArgs("Updated", 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := orm.UpdateRecordByID(bypassCtx(), "test.mock", 5, map[string]interface{}{"name": "Updated"}); err != nil {
		t.Fatal(err)
	}
}

func TestUnlinkWithMockDB(t *testing.T) {
	mock := setupMockORM(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "test_mock" WHERE \("id" = \$1\) FOR UPDATE`).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(9, "Drop"))
	mock.ExpectExec(`DELETE FROM "test_mock" WHERE \("id" = \$1\)`).
		WithArgs(9).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := orm.Unlink(bypassCtx(), "test.mock", 9); err != nil {
		t.Fatal(err)
	}
}

func TestSearchCountWithMockDB(t *testing.T) {
	mock := setupMockORM(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "test_mock" WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(11))

	n, err := orm.SearchCount(bypassCtx(), "test.mock", nil)
	if err != nil {
		t.Fatal(err)
	}
	if n != 11 {
		t.Fatalf("count = %d", n)
	}
}

func TestCreateWithMockDB(t *testing.T) {
	mock := setupMockORM(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "test_mock"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectCommit()

	model := orm.NewStubModelForTest("test.mock", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char, Required: true},
	})
	id, err := orm.Create(bypassCtx(), model, map[string]interface{}{"name": "New"})
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Fatalf("create id = %d", id)
	}
}
