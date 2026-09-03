package orm_test

import (
	"testing"

	"sumeru/core/orm"
)

func TestPrepareValuesMinMaxValidation(t *testing.T) {
	m := orm.NewStubModelForTest("test.range", []orm.FieldDefinition{
		{Name: "rating", Type: orm.Integer, String: "Rating", Min: floatPtr(0), Max: floatPtr(5)},
	})
	_, err := orm.PrepareValues(m, map[string]interface{}{"rating": 10}, orm.WriteOpCreate, orm.PrepareOptions{StrictUnknown: true})
	if err == nil {
		t.Fatal("expected validation error for rating > max")
	}
	out, err := orm.PrepareValues(m, map[string]interface{}{"rating": 3}, orm.WriteOpCreate, orm.PrepareOptions{StrictUnknown: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["rating"] != 3 {
		t.Fatalf("got %#v", out["rating"])
	}
}

func floatPtr(f float64) *float64 { return &f }
