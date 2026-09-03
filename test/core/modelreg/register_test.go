package modelreg_test

import (
	"reflect"
	"testing"

	"sumeru/core/modelmeta"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)

func TestParseSelection(t *testing.T) {
	opts := modelreg.ParseSelectionForTest("draft:Draft,done:Done")
	if len(opts) != 2 {
		t.Fatalf("got %d options", len(opts))
	}
	if opts[0][0] != "draft" || opts[0][1] != "Draft" {
		t.Fatalf("first option: %v", opts[0])
	}
	if opts[1][0] != "done" || opts[1][1] != "Done" {
		t.Fatalf("second option: %v", opts[1])
	}

	auto := modelreg.ParseSelectionForTest("low,high")
	if len(auto) != 2 || auto[0][1] != "Low" || auto[1][1] != "High" {
		t.Fatalf("auto labels: %v", auto)
	}

	if len(modelreg.ParseSelectionForTest(",  , ")) != 0 {
		t.Fatal("expected empty for blank parts")
	}
}

func TestParseDefault(t *testing.T) {
	if modelreg.ParseDefaultForTest("true", orm.Boolean) != true {
		t.Fatal("boolean true")
	}
	if modelreg.ParseDefaultForTest("no", orm.Boolean) != false {
		t.Fatal("boolean false")
	}
	if modelreg.ParseDefaultForTest("42", orm.Integer) != int64(42) {
		t.Fatal("integer")
	}
	if modelreg.ParseDefaultForTest("not-a-number", orm.Integer) != "not-a-number" {
		t.Fatal("integer fallback")
	}
	if modelreg.ParseDefaultForTest("3.5", orm.Float) != 3.5 {
		t.Fatal("float")
	}
	if modelreg.ParseDefaultForTest("hello", orm.Char) != "hello" {
		t.Fatal("string default")
	}
}

func TestMapMarkerType(t *testing.T) {
	ft, widget, err := modelreg.MapMarkerTypeForTest("Money")
	if err != nil || ft != orm.Numeric || widget != "monetary" {
		t.Fatalf("Money: ft=%s widget=%q err=%v", ft, widget, err)
	}
	_, _, err = modelreg.MapMarkerTypeForTest("NotARealMarker")
	if err == nil {
		t.Fatal("expected error for unknown marker")
	}
}

func TestResolveComodel(t *testing.T) {
	ctx := modelreg.NewRegisterCtxForTest().
		SetTypeMapping("sumeru/addons/base/models.CorePartner", "core.partner", "CorePartner", reflect.TypeOf(struct{}{}))

	comodel, err := ctx.ResolveComodel(reflect.TypeOf(modelmeta.Many2One[struct{}]{}), modelmeta.FieldTags{
		Comodel: "explicit.model",
	})
	if err != nil || comodel != "explicit.model" {
		t.Fatalf("explicit comodel: %q err=%v", comodel, err)
	}

	_, err = ctx.ResolveComodel(reflect.TypeOf(modelmeta.Many2One[modelmeta.Any]{}), modelmeta.FieldTags{})
	if err == nil {
		t.Fatal("expected error for Any without comodel tag")
	}

	fallback, err := ctx.ResolveComodel(reflect.TypeOf(modelmeta.Many2One[struct{}]{}), modelmeta.FieldTags{})
	if err != nil || fallback != "struct {}" {
		t.Fatalf("fallback comodel: %q err=%v", fallback, err)
	}
}

func TestIsEmbeddedModelMeta(t *testing.T) {
	type withMeta struct {
		modelmeta.ModelMeta
	}
	st := reflect.TypeOf(withMeta{})
	if !modelmeta.IsEmbeddedModelMeta(st.Field(0)) {
		t.Fatal("expected embedded ModelMeta")
	}
	if modelmeta.IsEmbeddedModelMeta(st.Field(0)) && st.NumField() == 1 {
		// only one field
	}
	type withName struct {
		Name string
	}
	st2 := reflect.TypeOf(withName{})
	if modelmeta.IsEmbeddedModelMeta(st2.Field(0)) {
		t.Fatal("regular field should not match")
	}
}
