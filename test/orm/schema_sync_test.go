package orm_test

import (
	"testing"

	"sumeru/core/orm"
)

func TestColumnTypeSQL(t *testing.T) {
	_, ok := orm.ColumnTypeSQL(orm.FieldDefinition{Name: "x", Type: orm.FieldType("nope")})
	if ok {
		t.Fatal("unknown type should return ok=false")
	}
	s, ok := orm.ColumnTypeSQL(orm.FieldDefinition{Name: "m", Type: orm.Float64})
	if !ok || s != "DOUBLE PRECISION" {
		t.Fatalf("float64: got %q %v", s, ok)
	}
	s, ok = orm.ColumnTypeSQL(orm.FieldDefinition{Name: "m", Type: orm.Float})
	if !ok || s != "REAL" {
		t.Fatalf("float: got %q %v", s, ok)
	}
}

func TestFormatAddColumnDefinitionBooleanDefault(t *testing.T) {
	got := orm.FormatAddColumnDefinition(orm.FieldDefinition{Name: "b", Type: orm.Boolean, DefaultVal: true}, "BOOLEAN")
	if got != "BOOLEAN NOT NULL DEFAULT TRUE" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatAddColumnDefinitionSkipsRuntimeUUIDDefault(t *testing.T) {
	got := orm.FormatAddColumnDefinition(orm.FieldDefinition{Name: "public_id", Type: orm.Char, DefaultVal: "uuid"}, "VARCHAR(36)")
	if got != "VARCHAR(36)" {
		t.Fatalf("got %q", got)
	}
}
