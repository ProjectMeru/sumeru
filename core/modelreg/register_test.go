package modelreg

import (
	"reflect"
	"testing"

	"sumeru/core/modelmeta"
	"sumeru/core/orm"
)

func TestParseSelection(t *testing.T) {
	opts := parseSelection("draft:Draft,done:Done")
	if len(opts) != 2 {
		t.Fatalf("got %d options", len(opts))
	}
	if opts[0][0] != "draft" || opts[0][1] != "Draft" {
		t.Fatalf("first option: %v", opts[0])
	}
	if opts[1][0] != "done" || opts[1][1] != "Done" {
		t.Fatalf("second option: %v", opts[1])
	}

	auto := parseSelection("low,high")
	if len(auto) != 2 || auto[0][1] != "Low" || auto[1][1] != "High" {
		t.Fatalf("auto labels: %v", auto)
	}

	if len(parseSelection(",  , ")) != 0 {
		t.Fatal("expected empty for blank parts")
	}
}

func TestParseDefault(t *testing.T) {
	if parseDefault("true", orm.Boolean) != true {
		t.Fatal("boolean true")
	}
	if parseDefault("no", orm.Boolean) != false {
		t.Fatal("boolean false")
	}
	if parseDefault("42", orm.Integer) != int64(42) {
		t.Fatal("integer")
	}
	if parseDefault("not-a-number", orm.Integer) != "not-a-number" {
		t.Fatal("integer fallback")
	}
	if parseDefault("3.5", orm.Float) != 3.5 {
		t.Fatal("float")
	}
	if parseDefault("hello", orm.Char) != "hello" {
		t.Fatal("string default")
	}
}

func TestMapMarkerType(t *testing.T) {
	ft, widget, err := mapMarkerType("Money")
	if err != nil || ft != orm.Numeric || widget != "monetary" {
		t.Fatalf("Money: ft=%s widget=%q err=%v", ft, widget, err)
	}
	_, _, err = mapMarkerType("NotARealMarker")
	if err == nil {
		t.Fatal("expected error for unknown marker")
	}
}

func TestResolveComodel(t *testing.T) {
	ctx := &registerCtx{
		typeNames: map[string]string{
			"sumeru/addons/base/models.CorePartner": "core.partner",
		},
		byName: map[string]reflect.Type{
			"CorePartner": reflect.TypeOf(struct{}{}),
		},
	}

	comodel, err := ctx.resolveComodel(reflect.TypeOf(modelmeta.Many2One[struct{}]{}), modelmeta.FieldTags{
		Comodel: "explicit.model",
	})
	if err != nil || comodel != "explicit.model" {
		t.Fatalf("explicit comodel: %q err=%v", comodel, err)
	}

	_, err = ctx.resolveComodel(reflect.TypeOf(modelmeta.Many2One[modelmeta.Any]{}), modelmeta.FieldTags{})
	if err == nil {
		t.Fatal("expected error for Any without comodel tag")
	}

	fallback, err := ctx.resolveComodel(reflect.TypeOf(modelmeta.Many2One[struct{}]{}), modelmeta.FieldTags{})
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
