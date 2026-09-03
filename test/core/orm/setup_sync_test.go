package orm_test

import (
	"slices"
	"testing"

	"sumeru/core/orm"
)

func TestModelsForModuleSchemaSyncBaseIncludesTranslation(t *testing.T) {
	orm.RegisterModelWithModule(
		orm.NewStubModelForTest("sys.module", []orm.FieldDefinition{{Name: "name", Type: orm.Char}}),
		"base",
	)
	orm.RegisterModelWithModule(
		orm.NewStubModelForTest("sys.translation", []orm.FieldDefinition{{Name: "lang", Type: orm.Char}}),
		"base",
	)
	t.Cleanup(func() {
		delete(orm.Registry, "sys.module")
		delete(orm.Registry, "sys.translation")
	})

	names, scoped := orm.ModelsForModuleSchemaSync("base")
	if !scoped {
		t.Fatal("expected scoped sync for base")
	}
	if !slices.Contains(names, "sys.translation") {
		t.Fatalf("expected sys.translation in base schema sync, got %v", names)
	}
}
