package models_test

import (
	"testing"

	_ "sumeru/addons/currency"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)

// TestCurrencyModelsRegistered verifies the SUM-PLAT-18 currency models are
// registered by the currency addon with the expected field shapes.
func TestCurrencyModelsRegistered(t *testing.T) {
	_ = modelreg.ActivateAll([]string{"base", "currency"})
	for _, name := range []string{"core.currency", "core.currency.rate"} {
		if _, ok := orm.Registry[name]; !ok {
			t.Fatalf("model %s not registered", name)
		}
	}

	currency, _ := orm.Registry["core.currency"]
	hasName := false
	for _, f := range currency.Fields() {
		if f.Name == "name" && f.Type == orm.Char {
			hasName = true
		}
	}
	if !hasName {
		t.Fatal("core.currency.name missing")
	}

	rate, _ := orm.Registry["core.currency.rate"]
	fields := map[string]orm.FieldDefinition{}
	for _, f := range rate.Fields() {
		fields[f.Name] = f
	}
	if f, ok := fields["currency_id"]; !ok || f.Type != orm.Many2One || f.Relation != "core.currency" {
		t.Fatalf("core.currency.rate.currency_id = %+v, want many2one core.currency", fields["currency_id"])
	}
	if f, ok := fields["rate"]; !ok || f.Type != orm.Float64 {
		t.Fatalf("core.currency.rate.rate = %+v, want float64", fields["rate"])
	}
	if f, ok := fields["date_from"]; !ok || f.Type != orm.Date {
		t.Fatalf("core.currency.rate.date_from = %+v, want date", fields["date_from"])
	}
}
