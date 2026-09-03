package modelreg_test

import (
	"reflect"
	"testing"

	"sumeru/core/modelmeta"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)

type inheritBase struct {
	modelmeta.ModelMeta `sumeru:"model=test.inherit.base"`

	Name modelmeta.String `sumeru:"string=Name"`
}

type inheritExtend struct {
	modelmeta.ModelMeta `sumeru:"inherit=test.inherit.base"`

	Score modelmeta.Integer `sumeru:"string=Score,default=0"`
}

func TestMustRegisterInheritMergesFields(t *testing.T) {
	const baseModel = "test.inherit.base"
	modelreg.ResetActivationForTest()
	prev := orm.Registry[baseModel]
	defer func() {
		if prev == nil {
			delete(orm.Registry, baseModel)
		} else {
			orm.Registry[baseModel] = prev
		}
	}()

	modelreg.MustRegister("test_inherit_base", &inheritBase{})
	modelreg.MustRegister("test_inherit_extend", &inheritExtend{})
	if err := modelreg.ActivateAll(nil); err != nil {
		t.Fatal(err)
	}

	m := orm.RegistryModel(baseModel)
	if m == nil {
		t.Fatal("base model missing from registry")
	}
	names := fieldNames(m.Fields())
	if len(names) != 2 {
		t.Fatalf("expected 2 fields, got %v", names)
	}
	if names[0] != "name" || names[1] != "score" {
		t.Fatalf("unexpected fields %v", names)
	}
	extended := orm.ModelsExtendedByModule("test_inherit_extend")
	if len(extended) != 1 || extended[0] != baseModel {
		t.Fatalf("ModelsExtendedByModule = %v, want [%s]", extended, baseModel)
	}
}

func TestMustRegisterInheritDuplicateFieldPanics(t *testing.T) {
	const baseModel = "test.inherit.dup"
	modelreg.ResetActivationForTest()
	prev := orm.Registry[baseModel]
	defer func() {
		if prev == nil {
			delete(orm.Registry, baseModel)
		} else {
			orm.Registry[baseModel] = prev
		}
	}()

	type dupBase struct {
		modelmeta.ModelMeta `sumeru:"model=test.inherit.dup"`
		Name                modelmeta.String
	}
	type dupExtend struct {
		modelmeta.ModelMeta `sumeru:"inherit=test.inherit.dup"`
		Name                modelmeta.String
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for duplicate field")
		}
	}()
	modelreg.MustRegister("test_inherit_dup_base", &dupBase{})
	modelreg.MustRegister("test_inherit_dup_extend", &dupExtend{})
	_ = modelreg.ActivateAll(nil)
}

func TestMustRegisterInheritUnknownTargetPanics(t *testing.T) {
	modelreg.ResetActivationForTest()
	type orphanExtend struct {
		modelmeta.ModelMeta `sumeru:"inherit=test.inherit.missing"`
		Flag                modelmeta.Boolean
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unknown inherit target")
		}
	}()
	modelreg.MustRegister("test_inherit_orphan", &orphanExtend{})
	_ = modelreg.ActivateAll(nil)
}

func TestModelSpecFromStructInherit(t *testing.T) {
	spec, err := modelmeta.ModelSpecFromStruct(reflect.TypeOf(inheritExtend{}))
	if err != nil {
		t.Fatal(err)
	}
	if !spec.Extend || spec.Name != "test.inherit.base" {
		t.Fatalf("spec = %+v", spec)
	}
}

func fieldNames(fields []orm.FieldDefinition) []string {
	var out []string
	for _, f := range fields {
		out = append(out, f.Name)
	}
	return out
}
