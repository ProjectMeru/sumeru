//go:build integration

// Package base_sync integration-tests the module data sync path (Addon.SyncToDB)
// for the base addon in isolation from other integration test binaries, because
// activating kernel models changes access-check behavior for tests that search
// sys.* without a bypass context.
package base_sync_test

import (
	"context"
	"database/sql"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	_ "sumeru/addons/base"
	"sumeru/core/modelreg"
	"sumeru/core/module"
	"sumeru/core/orm"
	_ "sumeru/core/ormmodels"
	"sumeru/test/harness"
)

// freshSyncTestDB points the ORM at a dedicated database that is dropped and
// recreated on every run, so the sync test never depends on (or disturbs)
// the shared test database and its assertions stay deterministic.
func freshSyncTestDB(t *testing.T, dsn string) string {
	t.Helper()
	const syncDBName = "sumeru_test_base_sync"
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open admin connection: %v", err)
	}
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + syncDBName); err != nil {
		t.Fatalf("drop %s: %v", syncDBName, err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + syncDBName); err != nil {
		t.Fatalf("create %s: %v", syncDBName, err)
	}
	_ = admin.Close()

	re := regexp.MustCompile(`\bdbname=\S+`)
	if !re.MatchString(dsn) {
		t.Fatalf("DSN has no dbname: %s", dsn)
	}
	return re.ReplaceAllString(dsn, "dbname="+syncDBName)
}

// TestBaseAddonSyncToDBLoadsCurrencyUomData exercises Addon.SyncToDB with the
// base manifest XML: currencies, rates, uom categories/units, views, actions,
// and menus must all land in the database, with rate/uom records resolving
// cross-record refs and natural keys (regression test for the
// core.currency.rate / uom.uom recordSyncSpecs entries).
func TestBaseAddonSyncToDBLoadsCurrencyUomData(t *testing.T) {
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDB(freshSyncTestDB(t, dsn))
	_ = modelreg.ActivateAll([]string{"base"})
	if err := orm.SyncRegistrySchemaForNames(orm.InitialSetupModelNames); err != nil {
		t.Fatalf("kernel schema sync: %v", err)
	}
	if err := orm.SyncRegistrySchemaForModule("base"); err != nil {
		t.Fatalf("base schema sync: %v", err)
	}
	root := harness.RepoRoot(t)
	discovered, err := module.DiscoverAddonRoots([]string{filepath.Join(root, "addons")})
	if err != nil {
		t.Fatalf("discover addons: %v", err)
	}
	addon, ok := discovered["base"]
	if !ok {
		t.Fatal("base addon not discovered")
	}
	ctx := orm.ContextWithBypass(context.Background(), true)
	if err := addon.SyncToDB(ctx); err != nil {
		t.Fatalf("base SyncToDB: %v", err)
	}

	counts := map[string]struct {
		model string
		want  int
	}{
		"currencies":     {"core.currency", 8},
		"rates":          {"core.currency.rate", 8},
		"uom categories": {"uom.category", 5},
		"uom units":      {"uom.uom", 16},
	}
	for label, c := range counts {
		n, err := orm.SearchCount(ctx, c.model, nil)
		if err != nil {
			t.Fatalf("count %s: %v", label, err)
		}
		if n != c.want {
			t.Fatalf("%s: got %d rows, want %d", label, n, c.want)
		}
	}

	// Rate rows resolve the currency ref and carry the expected value.
	eur, err := orm.SearchOne(ctx, "core.currency", map[string]interface{}{"name": "EUR"})
	if err != nil {
		t.Fatalf("find EUR: %v", err)
	}
	eurID, _ := orm.CoerceInt64(eur["id"])
	rateRow, err := orm.SearchOne(ctx, "core.currency.rate", map[string]interface{}{
		"currency_id": int(eurID), "date_from": "2026-01-01",
	})
	if err != nil {
		t.Fatalf("find EUR rate: %v", err)
	}
	if rate, ok := rateRow["rate"].(float64); !ok || math.Abs(rate-1.09) > 1e-9 {
		t.Fatalf("EUR rate = %v, want 1.09", rateRow["rate"])
	}

	// UoM rows resolve the category ref and carry the expected factor.
	kg, err := orm.SearchOne(ctx, "uom.uom", map[string]interface{}{"name": "kg"})
	if err != nil {
		t.Fatalf("find kg: %v", err)
	}
	if f, ok := kg["factor"].(float64); !ok || math.Abs(f-1) > 1e-9 {
		t.Fatalf("kg factor = %v, want 1", kg["factor"])
	}

	// Admin UI metadata: views, window actions, and menus for the new models.
	viewNames := []string{"view_core.currency.rate_list", "view_uom.uom_form"}
	for _, name := range viewNames {
		if _, err := orm.SearchOne(ctx, "sys.view", map[string]interface{}{"name": name}); err != nil {
			t.Fatalf("sys.view %q missing: %v", name, err)
		}
	}
	actionNames := []string{"Currencies", "Currency Rates", "UoM Categories", "Units of Measure"}
	for _, name := range actionNames {
		if _, err := orm.SearchOne(ctx, "sys.action.window", map[string]interface{}{"name": name}); err != nil {
			t.Fatalf("sys.action.window %q missing: %v", name, err)
		}
	}
	menus := []string{"Currencies & Units", "Currencies", "Currency Rates", "UoM Categories", "Units of Measure"}
	for _, menu := range menus {
		if _, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"name": menu}); err != nil {
			t.Fatalf("menu %q missing: %v", menu, err)
		}
	}
}
