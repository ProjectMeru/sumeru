package orm_test

import (
	"testing"
	"sumeru/core/orm"
)


func TestCorePartnerContactsExtension(t *testing.T) {
	m := orm.RegistryModel("core.partner")
	if m == nil {
		t.Fatal("core.partner not registered")
	}
	found := false
	for _, f := range m.Fields() {
		if f.Name == "customer_rank" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("customer_rank field missing from core.partner after contacts extension")
	}
	extended := orm.ModelsExtendedByModule("contacts")
	if len(extended) != 1 || extended[0] != "core.partner" {
		t.Fatalf("ModelsExtendedByModule(contacts) = %v", extended)
	}
}
