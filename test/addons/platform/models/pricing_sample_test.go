package models_test

import (
	"testing"

	_ "sumeru/addons/currency"
	_ "sumeru/addons/platform"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)

func TestPlatformPricingSampleMoneyField(t *testing.T) {
	_ = modelreg.ActivateAll([]string{"base", "currency", "platform"})
	inst, ok := orm.Registry["platform.pricing.sample"]
	if !ok {
		t.Fatal("platform.pricing.sample not registered")
	}
	fields := map[string]orm.FieldDefinition{}
	for _, f := range inst.Fields() {
		fields[f.Name] = f
	}
	amount, ok := fields["amount"]
	if !ok || amount.Widget != "monetary" || amount.Currency != "currency_id" {
		t.Fatalf("amount field = %+v, want monetary with currency_id", amount)
	}
}
