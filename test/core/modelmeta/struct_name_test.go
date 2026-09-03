package modelmeta_test

import (
	"reflect"
	"testing"
	"sumeru/core/modelmeta"
)


func TestModelNameFromStruct(t *testing.T) {
	type tagged struct {
		modelmeta.ModelMeta `sumeru:"model=core.company"`
	}
	name, err := modelmeta.ModelNameFromStruct(reflect.TypeOf(tagged{}))
	if err != nil {
		t.Fatal(err)
	}
	if name != "core.company" {
		t.Fatalf("got %q want core.company", name)
	}
}
