//go:build integration

// Package core_sync integration-tests the module data sync path (Addon.SyncToDB)
// for the core addon install closure (base, currency, uom) in isolation from
// other integration test binaries, because activating kernel models changes
// access-check behavior for tests that search sys.* without a bypass context.
package core_sync_test

import (
	"context"
	"database/sql"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	_ "sumeru/addons/base"
	_ "sumeru/addons/currency"
	_ "sumeru/addons/uom"
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

// TestCoreAddonSyncToDBLoadsCurrencyUomData exercises Addon.SyncToDB across the
// core addon install closure (base, currency, uom): the currency catalog (with
// no seeded rates — rate rows are business data created by verticals/providers),
// uom categories/units, views, actions, and menus must all land in the
// database, with uom records resolving cross-record refs and natural keys
// (regression test for the core.currency.rate / uom.uom recordSyncSpecs
// entries).
func TestCoreAddonSyncToDBLoadsCurrencyUomData(t *testing.T) {
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDB(freshSyncTestDB(t, dsn))
	_ = modelreg.ActivateAll([]string{"base", "currency", "uom"})
	// Sync the kernel setup subset, skipping models owned by addons this
	// binary does not load (e.g. sys.translation from the i18n addon) —
	// the core data sync does not need them.
	var kernelModels []string
	for _, name := range orm.InitialSetupModelNames {
		if _, ok := orm.Registry[name]; ok {
			kernelModels = append(kernelModels, name)
		}
	}
	if err := orm.SyncRegistrySchemaForNames(kernelModels); err != nil {
		t.Fatalf("kernel schema sync: %v", err)
	}
	root := harness.RepoRoot(t)
	discovered, err := module.DiscoverAddonRoots([]string{filepath.Join(root, "addons")})
	if err != nil {
		t.Fatalf("discover addons: %v", err)
	}
	ctx := orm.ContextWithBypass(context.Background(), true)
	// Install-closure order: schema sync then data sync per module.
	for _, name := range []string{"base", "currency", "uom"} {
		addon, ok := discovered[name]
		if !ok {
			t.Fatalf("addon %s not discovered", name)
		}
		if err := orm.SyncRegistrySchemaForModule(name); err != nil {
			t.Fatalf("%s schema sync: %v", name, err)
		}
		if err := addon.SyncToDB(ctx); err != nil {
			t.Fatalf("%s SyncToDB: %v", name, err)
		}
	}

	counts := map[string]struct {
		model string
		want  int
	}{
		"currencies":     {"core.currency", 8},
		"rates":          {"core.currency.rate", 0}, // no core seed: filled by verticals/providers
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

	// The currency catalog loads with its symbol data.
	eur, err := orm.SearchOne(ctx, "core.currency", map[string]interface{}{"name": "EUR"})
	if err != nil {
		t.Fatalf("find EUR: %v", err)
	}
	if sym, ok := eur["symbol"].(string); !ok || sym != "€" {
		t.Fatalf("EUR symbol = %v, want €", eur["symbol"])
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
	// Section menus live in their own module (currency / uom) under the base
	// Settings root.
	menus := []string{
		"Currencies", "All Currencies", "Currency Rates",
		"Units of Measure", "UoM Categories", "UoM Units",
	}
	for _, menu := range menus {
		if _, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"name": menu}); err != nil {
			t.Fatalf("menu %q missing: %v", menu, err)
		}
	}
}
