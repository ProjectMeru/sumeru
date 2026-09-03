package orm_test

import (
	"testing"
	"sumeru/core/orm"
)


func TestModelsForModuleSchemaSyncIncludesExtensions(t *testing.T) {
	orm.RecordModelExtendedBy("core.partner", "contacts")
	names, scoped := orm.ModelsForModuleSchemaSync("contacts")
	if !scoped {
		t.Fatal("expected scoped sync for contacts")
	}
	found := false
	for _, name := range names {
		if name == "core.partner" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected core.partner in sync list, got %v", names)
	}
}
