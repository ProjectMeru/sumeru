//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/sdk"
)

type suiteItem struct {
	sdk.Model   `sumeru:"model=test.suite.item"`
	Name        sdk.String            `sumeru:"string=Name"`
	Kind        sdk.String            `sumeru:"string=Kind"`
	Amount      sdk.Float             `sumeru:"string=Amount"`
	StoredTotal sdk.Float             `sumeru:"compute=storedTotal,store,string=Stored total"`
	State       sdk.Selection[string] `sumeru:"status=applySuiteStatus,selection=draft:Draft,done:Done,string=State,default=draft"`
}

func init() {
	modelreg.MustRegister("testsuite", &suiteItem{})
}

var (
	computeBroken bool
	statusBroken  bool
)

func storedTotal(_ context.Context, rec map[string]interface{}) (any, error) {
	if computeBroken {
		return "bad", nil
	}
	amount, _ := sdk.CoerceFloat64(rec["amount"])
	return amount * 2, nil
}

func applySuiteStatus(_ context.Context, rec map[string]interface{}) (any, error) {
	if statusBroken {
		return "nope", nil
	}
	if rec["state"] == "done" {
		return "done", nil
	}
	return rec["state"], nil
}

func setupSuite(t *testing.T) context.Context {
	t.Helper()
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDB(dsn)
	if err := modelreg.ActivateAll(nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := orm.Registry["test.suite.item"]; !ok {
		t.Fatal("test.suite.item is not registered")
	}
	sdk.BindCompute("test.suite.item", "stored_total", storedTotal)
	sdk.BindStatus("test.suite.item", "state", applySuiteStatus)
	if err := orm.SyncRegistrySchemaForNames([]string{"test.suite.item"}); err != nil {
		t.Fatalf("schema sync: %v", err)
	}
	computeBroken = false
	statusBroken = false
	return orm.ContextWithBypass(context.Background(), true)
}

func readItem(t *testing.T, ctx context.Context, id int) map[string]interface{} {
	t.Helper()
	row, err := sdk.SearchOne(ctx, "test.suite.item", map[string]interface{}{"id": id})
	if err != nil {
		t.Fatal(err)
	}
	return row
}

func TestSuiteCRUD(t *testing.T) {
	ctx := setupSuite(t)
	id, err := sdk.Create(ctx, "test.suite.item", map[string]interface{}{"name": "alpha", "kind": "a", "amount": 1})
	if err != nil {
		t.Fatal(err)
	}
	if readItem(t, ctx, id)["name"] != "alpha" {
		t.Fatalf("name = %#v", readItem(t, ctx, id)["name"])
	}
	if _, err := sdk.Write(ctx, "test.suite.item", []int{id}, map[string]interface{}{"name": "beta"}); err != nil {
		t.Fatal(err)
	}
	if readItem(t, ctx, id)["name"] != "beta" {
		t.Fatalf("name = %#v", readItem(t, ctx, id)["name"])
	}
	if _, err := sdk.Unlink(ctx, "test.suite.item", []int{id}); err != nil {
		t.Fatal(err)
	}
	rows, err := sdk.Search(ctx, "test.suite.item", [][]interface{}{{"id", "=", id}})
	if err != nil {
		t.Fatal(err)
	}
	if 0 != len(rows) {
		t.Fatalf("unlinked row still readable: %#v", rows)
	}
}

func TestSuiteStatusAndCompute(t *testing.T) {
	ctx := setupSuite(t)
	id, err := sdk.Create(ctx, "test.suite.item", map[string]interface{}{
		"name": "status", "kind": "a", "amount": 2, "state": "draft",
	})
	if err != nil {
		t.Fatal(err)
	}
	row := readItem(t, ctx, id)
	total, ok := sdk.CoerceFloat64(row["stored_total"])
	if !ok || total != 4 {
		t.Fatalf("stored_total = %#v", row["stored_total"])
	}

	if _, err := sdk.Write(ctx, "test.suite.item", []int{id}, map[string]interface{}{"state": "done"}); err != nil {
		t.Fatal(err)
	}
	if readItem(t, ctx, id)["state"] != "done" {
		t.Fatalf("state = %#v", readItem(t, ctx, id)["state"])
	}

	statusBroken = true
	_, err = sdk.Write(ctx, "test.suite.item", []int{id}, map[string]interface{}{"state": "draft"})
	statusBroken = false
	if err == nil || !strings.Contains(err.Error(), "state") {
		t.Fatalf("bad stage: %v", err)
	}
	if readItem(t, ctx, id)["state"] != "done" {
		t.Fatalf("state changed after rejected write: %#v", readItem(t, ctx, id)["state"])
	}

	computeBroken = true
	_, err = sdk.Write(ctx, "test.suite.item", []int{id}, map[string]interface{}{"amount": 9})
	computeBroken = false
	if err == nil || !strings.Contains(err.Error(), "stored_total") {
		t.Fatalf("bad compute: %v", err)
	}
	row = readItem(t, ctx, id)
	amount, _ := sdk.CoerceFloat64(row["amount"])
	total, _ = sdk.CoerceFloat64(row["stored_total"])
	if amount != 2 || total != 4 {
		t.Fatalf("row changed after rejected compute: amount=%v stored_total=%v", amount, total)
	}
}

func TestSuiteReadGroup(t *testing.T) {
	ctx := setupSuite(t)
	kind := fmt.Sprintf("g%d", time.Now().UnixNano())
	for _, amount := range []float64{10, 5} {
		if _, err := sdk.Create(ctx, "test.suite.item", map[string]interface{}{
			"name": "group", "kind": kind, "amount": amount, "state": "draft",
		}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := orm.ReadGroup(ctx, "test.suite.item", [][]interface{}{{"kind", "=", kind}}, orm.ReadGroupSpec{
		GroupBy: []string{"kind"},
		Fields:  []orm.ReadGroupField{{Name: "amount", Field: "amount", Measure: "sum"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if 1 != len(rows) {
		t.Fatalf("rows = %#v", rows)
	}
	sum, ok := sdk.CoerceFloat64(rows[0]["amount"])
	if !ok || sum != 15 {
		t.Fatalf("sum = %#v", rows[0]["amount"])
	}
}
