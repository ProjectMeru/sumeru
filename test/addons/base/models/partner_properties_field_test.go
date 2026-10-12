package models_test

import (
	"testing"

	_ "sumeru/addons/base"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)

func TestCorePartnerExtraPropertiesField(t *testing.T) {
	_ = modelreg.ActivateAll([]string{"base"})
	inst, ok := orm.Registry["core.partner"]
	if !ok {
		t.Fatal("core.partner not registered")
	}
	for _, f := range inst.Fields() {
		if f.Name == "extra_properties" && f.Type == orm.Properties {
			return
		}
	}
	t.Fatal("core.partner.extra_properties properties field missing")
}
