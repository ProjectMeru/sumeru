package models_test

import (
	"testing"

	_ "sumeru/addons/uom"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)

// TestUomModelsRegistered verifies the SUM-PLAT-18 uom models are registered
// by the uom addon with the expected field shapes.
func TestUomModelsRegistered(t *testing.T) {
	_ = modelreg.ActivateAll([]string{"base", "uom"})
	for _, name := range []string{"uom.category", "uom.uom"} {
		if _, ok := orm.Registry[name]; !ok {
			t.Fatalf("model %s not registered", name)
		}
	}

	uom, _ := orm.Registry["uom.uom"]
	fields := map[string]orm.FieldDefinition{}
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
}
