//go:build integration

package form_test

import (
	"context"
	"os"
	"testing"

	"sumeru/core/orm"
	"sumeru/test/harness/form"
)

func TestFormSaveAndRecordIntegration(t *testing.T) {
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDBWithPool(dsn, orm.DBPoolSettings{MaxOpenConns: 5, MaxIdleConns: 2})
	if !orm.IsInitialized() {
		t.Skip("database not initialized (run setup or apply schema first)")
	}

	model := stubModel{
		name: "test.form.int",
		fields: []orm.FieldDefinition{
			{Name: "name", Type: orm.Char, Required: true},
			{Name: "qty", Type: orm.Integer, DefaultVal: 1},
		},
	}
	orm.Registry[model.name] = model

	ctx := orm.ContextWithBypass(context.Background(), true)
	if err := orm.SyncModels(); err != nil {
		t.Fatalf("SyncModels: %v", err)
	}

	f, err := form.New(ctx, model)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Set("name", "A"); err != nil {
		t.Fatal(err)
	}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	if f.ID() == 0 {
		t.Fatal("expected non-zero id after save")
	}

	rec, err := f.Record()
	if err != nil {
		t.Fatal(err)
	}
	if rec["name"] != "A" {
		t.Fatalf("record name = %v, want A", rec["name"])
	}
	if qty, ok := orm.CoerceInt64(rec["qty"]); !ok || qty != 1 {
		t.Fatalf("record qty = %v, want 1", rec["qty"])
	}
}
