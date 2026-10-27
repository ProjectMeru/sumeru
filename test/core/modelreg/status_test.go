package modelreg_test

import (
	"fmt"
	"strings"
	"testing"

	"sumeru/core/modelmeta"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
	testreg "sumeru/test/core/modelreg"
)

func TestStatusTagRegistersStatusbar(t *testing.T) {
	modelreg.ResetActivationForTest()
	orm.Registry = map[string]orm.Model{}
	modelreg.MustRegister("test", &testreg.StatusBarModel{})
	if err := modelreg.ActivateAll(nil); err != nil {
		t.Fatal(err)
	}
	m := orm.Registry["test.status.bar"]
	if m == nil {
		t.Fatal("missing model")
	}
	var state orm.FieldDefinition
	found := false
	for _, field := range m.Fields() {
		if field.Name == "state" {
			state = field
			found = true
		}
	}
	if !found {
		t.Fatal("missing state")
	}
	if !state.Status {
		t.Fatal("status flag")
	}
	if state.Widget != "statusbar" {
		t.Fatalf("widget %q", state.Widget)
	}
	got := map[string]string{}
	for _, opt := range state.Selection {
		if len(opt) >= 2 {
			got[opt[0]] = opt[1]
		}
	}
	if got["draft"] != "Draft" || got["active"] != "Active" {
		t.Fatalf("stages %v", state.Selection)
	}
}

func TestStatusTagRejectsNonSelection(t *testing.T) {
	modelreg.ResetActivationForTest()
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(fmt.Sprint(recovered), "selection") {
			t.Fatal(recovered)
		}
	}()
	modelreg.MustRegister("test", &testreg.BadStatusModel{})
}

func TestStatusTagRejectsValue(t *testing.T) {
	_, err := modelmeta.ParseFieldTag("status=draft")
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatal(err)
	}
}
