package models_test

import (
	"context"
	"testing"
	"time"

	_ "sumeru/addons/base"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/sdk"
)

// TestCurrencyUomModelsRegistered verifies the SUM-PLAT-18 platform models are
// registered with the expected field shapes.
func TestCurrencyUomModelsRegistered(t *testing.T) {
	_ = modelreg.ActivateAll([]string{"base"})
	ctx := context.Background()
	for _, name := range []string{"core.currency", "core.currency.rate", "uom.category", "uom.uom"} {
		if _, ok := orm.Registry[name]; !ok {
			t.Fatalf("model %s not registered", name)
		}
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

	uom, _ := orm.Registry["uom.uom"]
	fields = map[string]orm.FieldDefinition{}
	for _, f := range uom.Fields() {
		fields[f.Name] = f
	}
	if f, ok := fields["category_id"]; !ok || f.Type != orm.Many2One || f.Relation != "uom.category" {
		t.Fatalf("uom.uom.category_id = %+v, want many2one uom.category", fields["category_id"])
	}
	if f, ok := fields["factor"]; !ok || f.Type != orm.Float64 {
		t.Fatalf("uom.uom.factor = %+v, want float64", fields["factor"])
	}
	if f, ok := fields["rounding"]; !ok || f.Type != orm.Float64 {
		t.Fatalf("uom.uom.rounding = %+v, want float64", fields["rounding"])
	}

	// The currency tag must keep flowing from struct tag to field definition.
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

	// ConvertMoneyField against a real base model field without currency tag.
	_, err := sdk.ConvertMoneyField(ctx, "core.user", "name", map[string]interface{}{"name": "x"}, 1, time.Time{})
	if err == nil {
		t.Fatal("expected error: core.user.name is not currency-tagged")
	}
}
