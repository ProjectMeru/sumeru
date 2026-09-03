package orm_test

import (
	"strings"
	"testing"

	"sumeru/core/orm"
)

type prepModel struct {
	name   string
	fields []orm.FieldDefinition
}

func (m prepModel) ModelName() string         { return m.name }
func (m prepModel) Fields() []orm.FieldDefinition { return m.fields }

func TestPrepareValues_rejectsUnknownOnCreate(t *testing.T) {
	m := prepModel{
		name: "test.prep",
		fields: []orm.FieldDefinition{
			{Name: "name", Type: orm.Char, Required: true},
		},
	}
	_, err := orm.PrepareValues(m, map[string]interface{}{
		"name":    "ok",
		"evilcol": "x",
	}, orm.WriteOpCreate, orm.PrepareOptions{StrictUnknown: true})
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestPrepareValues_whitelistsAndRequires(t *testing.T) {
	m := prepModel{
		name: "test.prep",
		fields: []orm.FieldDefinition{
			{Name: "name", Type: orm.Char, Required: true},
			{Name: "active", Type: orm.Boolean, Required: true, DefaultVal: true},
		},
	}
	out, err := orm.PrepareValues(m, map[string]interface{}{"name": "a"}, orm.WriteOpCreate, orm.PrepareOptions{StrictUnknown: true})
	if err != nil {
		t.Fatal(err)
	}
	if out["name"] != "a" {
		t.Fatalf("name=%v", out["name"])
	}
	if out["active"] != true {
		t.Fatalf("active default=%v", out["active"])
	}
}

func TestPrepareValues_writeDropsUnknown(t *testing.T) {
	m := prepModel{
		name: "test.prep",
		fields: []orm.FieldDefinition{
			{Name: "name", Type: orm.Char},
		},
	}
	out, err := orm.PrepareValues(m, map[string]interface{}{
		"name": "n",
		"skip": 1,
	}, orm.WriteOpWrite, orm.PrepareOptions{StrictUnknown: false})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := out["skip"]; ok {
		t.Fatal("unknown key should be dropped")
	}
}
