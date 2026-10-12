//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	_ "sumeru/addons/base"
	_ "sumeru/addons/currency"
	_ "sumeru/addons/i18n"
	_ "sumeru/addons/uom"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/sdk"
)

const rateDateLayout = "2006-01-02"

func setupCurrencyUom(t *testing.T) context.Context {
	t.Helper()
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDB(dsn)
	// i18n is loaded so ensureExtraIndexes can resolve sys.translation on an
	// already-installed database (the index is created for any synced schema).
	_ = modelreg.ActivateAll([]string{"base", "i18n", "currency", "uom"})
	if err := orm.SyncRegistrySchemaForNames([]string{
		"core.currency", "core.currency.rate", "uom.category", "uom.uom",
	}); err != nil {
		t.Fatalf("schema sync: %v", err)
	}
	return orm.ContextWithBypass(context.Background(), true)
}

// runSuffix keeps record names unique per test run so repeated runs against a
// shared database never collide on core.currency.name / uom.category.name.
var runSuffix = fmt.Sprintf("%d", time.Now().UnixNano())

func createCurrency(t *testing.T, ctx context.Context, name, symbol string) int {
	t.Helper()
	id, err := orm.Create(ctx, orm.Registry["core.currency"], map[string]interface{}{
		"name": name + runSuffix, "symbol": symbol, "active": true,
	})
	if err != nil {
		t.Fatalf("create currency %s: %v", name, err)
	}
	t.Cleanup(func() { _ = orm.Unlink(ctx, "core.currency", id) })
	return id
}

func createRate(t *testing.T, ctx context.Context, currencyID int, rate float64, dateFrom string) int {
	t.Helper()
	id, err := orm.Create(ctx, orm.Registry["core.currency.rate"], map[string]interface{}{
		"currency_id": currencyID, "rate": rate, "date_from": dateFrom, "active": true,
	})
	if err != nil {
		t.Fatalf("create rate for %d: %v", currencyID, err)
	}
	t.Cleanup(func() { _ = orm.Unlink(ctx, "core.currency.rate", id) })
	return id
}

func createUom(t *testing.T, ctx context.Context, name string, categoryID int, factor, rounding float64) int {
	t.Helper()
	id, err := orm.Create(ctx, orm.Registry["uom.uom"], map[string]interface{}{
		"name": name + runSuffix, "category_id": categoryID, "factor": factor,
		"rounding": rounding, "active": true,
	})
	if err != nil {
		t.Fatalf("create uom %s: %v", name, err)
	}
	t.Cleanup(func() { _ = orm.Unlink(ctx, "uom.uom", id) })
	return id
}

func createUomCategory(t *testing.T, ctx context.Context, name string) int {
	t.Helper()
	id, err := orm.Create(ctx, orm.Registry["uom.category"], map[string]interface{}{"name": name + runSuffix})
	if err != nil {
		t.Fatalf("create category %s: %v", name, err)
	}
	t.Cleanup(func() { _ = orm.Unlink(ctx, "uom.category", id) })
	return id
}

func TestIntegrationConvertCurrency(t *testing.T) {
	ctx := setupCurrencyUom(t)
	usd := createCurrency(t, ctx, "TUSD", "$")
	eur := createCurrency(t, ctx, "TEUR", "€")
	idr := createCurrency(t, ctx, "TIDR", "Rp")
	createRate(t, ctx, usd, 1.0, "2026-01-01")
	createRate(t, ctx, eur, 1.09, "2026-01-01")
	createRate(t, ctx, idr, 0.000064, "2026-01-01")
	date := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	got, err := sdk.ConvertCurrency(ctx, 100, usd, eur, date)
	if err != nil {
		t.Fatalf("USD->EUR: %v", err)
	}
	if math.Abs(got-91.7431192660) > 1e-6 {
		t.Fatalf("USD->EUR = %v, want ~91.7431", got)
	}

	got, err = sdk.ConvertCurrency(ctx, 100, eur, usd, date)
	if err != nil {
		t.Fatalf("EUR->USD: %v", err)
	}
	if math.Abs(got-109) > 1e-6 {
		t.Fatalf("EUR->USD = %v, want 109", got)
	}

	got, err = sdk.ConvertCurrency(ctx, 100, usd, idr, date)
	if err != nil {
		t.Fatalf("USD->IDR: %v", err)
	}
	if math.Abs(got-1562500) > 1e-6 {
		t.Fatalf("USD->IDR = %v, want 1562500", got)
	}
}

func TestIntegrationConvertCurrencyUsesNewestRate(t *testing.T) {
	ctx := setupCurrencyUom(t)
	usd := createCurrency(t, ctx, "TUSD2", "$")
	eur := createCurrency(t, ctx, "TEUR2", "€")
	createRate(t, ctx, usd, 1.0, "2026-01-01")
	createRate(t, ctx, eur, 1.09, "2026-01-01")
	createRate(t, ctx, eur, 1.15, "2026-06-01")

	before, err := sdk.ConvertCurrency(ctx, 100, usd, eur, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("before new rate: %v", err)
	}
	after, err := sdk.ConvertCurrency(ctx, 100, usd, eur, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("after new rate: %v", err)
	}
	if math.Abs(before-91.7431192660) > 1e-6 {
		t.Fatalf("before = %v, want ~91.7431", before)
	}
	if math.Abs(after-86.9565217391) > 1e-6 {
		t.Fatalf("after = %v, want ~86.9565", after)
	}
}

func TestIntegrationConvertCurrencyMissingRate(t *testing.T) {
	ctx := setupCurrencyUom(t)
	usd := createCurrency(t, ctx, "TUSD3", "$")
	orphan := createCurrency(t, ctx, "TORPHAN", "O")
	createRate(t, ctx, usd, 1.0, "2026-01-01")

	_, err := sdk.ConvertCurrency(ctx, 100, usd, orphan, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected error: no rate for orphan currency")
	}
}

func TestIntegrationConvertUom(t *testing.T) {
	ctx := setupCurrencyUom(t)
	catID := createUomCategory(t, ctx, "TWeight")
	kg := createUom(t, ctx, "tkg", catID, 1, 0.001)
	g := createUom(t, ctx, "tg", catID, 0.001, 0.001)
	ton := createUom(t, ctx, "tton", catID, 1000, 0.001)

	got, err := sdk.ConvertUom(ctx, 1, kg, g)
	if err != nil {
		t.Fatalf("kg->g: %v", err)
	}
	if math.Abs(got-1000) > 1e-9 {
		t.Fatalf("kg->g = %v, want 1000", got)
	}

	got, err = sdk.ConvertUom(ctx, 500, g, kg)
	if err != nil {
		t.Fatalf("g->kg: %v", err)
	}
	if math.Abs(got-0.5) > 1e-9 {
		t.Fatalf("g->kg = %v, want 0.5", got)
	}

	got, err = sdk.ConvertUom(ctx, 1.5, ton, kg)
	if err != nil {
		t.Fatalf("ton->kg: %v", err)
	}
	if math.Abs(got-1500) > 1e-9 {
		t.Fatalf("ton->kg = %v, want 1500", got)
	}
}

func TestIntegrationConvertUomCrossCategory(t *testing.T) {
	ctx := setupCurrencyUom(t)
	weightID := createUomCategory(t, ctx, "TWeight2")
	lengthID := createUomCategory(t, ctx, "TLength")
	kg := createUom(t, ctx, "tkg2", weightID, 1, 0.001)
	m := createUom(t, ctx, "tm", lengthID, 1, 0.01)

	if _, err := sdk.ConvertUom(ctx, 1, kg, m); err == nil {
		t.Fatal("expected error: cross-category conversion")
	}
}

func TestIntegrationConvertMoneyField(t *testing.T) {
	ctx := setupCurrencyUom(t)
	usd := createCurrency(t, ctx, "TUSD4", "$")
	eur := createCurrency(t, ctx, "TEUR4", "€")
	createRate(t, ctx, usd, 1.0, "2026-01-01")
	createRate(t, ctx, eur, 1.09, "2026-01-01")

	orm.RegisterStubModelForTest(t, "stub.int.money", []orm.FieldDefinition{
		{Name: "amount", Type: orm.Float64, Currency: "currency_id"},
	})
	rec := map[string]interface{}{"amount": 100.0, "currency_id": eur}
	got, err := sdk.ConvertMoneyField(ctx, "stub.int.money", "amount", rec, usd, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ConvertMoneyField: %v", err)
	}
	if math.Abs(got-109) > 1e-6 {
		t.Fatalf("ConvertMoneyField = %v, want 109 (100 EUR at 1.09 = 109 USD)", got)
	}
}
