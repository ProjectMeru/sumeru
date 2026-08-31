package modelmeta_test

import (
	"reflect"
	"testing"

	"sumeru/core/modelmeta"
	"sumeru/core/sdk"
)

type inheritTagged struct {
	sdk.Model `sumeru:"inherit=core.partner"`
	Extra     sdk.Integer
}

func TestModelSpecFromStructInherit(t *testing.T) {
	spec, err := modelmeta.ModelSpecFromStruct(reflect.TypeOf(inheritTagged{}))
	if err != nil {
		t.Fatal(err)
	}
	if !spec.Extend || spec.Name != "core.partner" {
		t.Fatalf("spec = %+v", spec)
	}
}

func TestParseModelTagInherit(t *testing.T) {
	name, err := modelmeta.ParseModelTag("inherit=core.partner")
	if err != nil {
		t.Fatal(err)
	}
	if name != "core.partner" {
		t.Fatalf("got %q", name)
	}
}

func TestParseModelTagMutuallyExclusive(t *testing.T) {
	_, err := modelmeta.ParseModelTag("model=core.partner,inherit=core.partner")
	if err == nil {
		t.Fatal("expected error when model= and inherit= are both set")
	}
}
