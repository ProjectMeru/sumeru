package form_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/orm"
	"sumeru/test/harness/form"
)

type stubModel struct {
	name   string
	fields []orm.FieldDefinition
}

func (m stubModel) ModelName() string             { return m.name }
func (m stubModel) Fields() []orm.FieldDefinition { return m.fields }

func mustNew(t *testing.T, m orm.Model) *form.Form {
	t.Helper()
	f, err := form.New(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func recInt(rec map[string]interface{}, key string) int {
	switch v := rec[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func TestDefaultsAndComputed(t *testing.T) {
	model := stubModel{
		name: "test.form.a",
		fields: []orm.FieldDefinition{
			{Name: "f1", Type: orm.Char, Required: true},
			{Name: "f2", Type: orm.Integer, DefaultVal: 42},
			{Name: "f3", Type: orm.Integer, Compute: "half", ComputeStore: false},
			{Name: "f4", Type: orm.Integer, Compute: "double", ComputeStore: false},
		},
	}
	orm.RegisterCompute(model.name, "f3", []string{"f2"}, func(_ context.Context, rec map[string]interface{}) (interface{}, error) {
		return recInt(rec, "f2") / 2, nil
	})
	orm.RegisterCompute(model.name, "f4", []string{"f2"}, func(_ context.Context, rec map[string]interface{}) (interface{}, error) {
		return recInt(rec, "f2") * 2, nil
	})

	f := mustNew(t, model)
	if got := f.Get("f2"); got != 42 {
		t.Fatalf("default f2 = %v, want 42", got)
	}
	if got := f.Get("f3"); got != 21 {
		t.Fatalf("computed f3 = %v, want 21", got)
	}
	if got := f.Get("f4"); got != 84 {
		t.Fatalf("computed f4 = %v, want 84", got)
	}

	if err := f.Set("f2", 8); err != nil {
		t.Fatal(err)
	}
	if got := f.Get("f3"); got != 4 {
		t.Fatalf("computed f3 = %v, want 4", got)
	}
	if got := f.Get("f4"); got != 16 {
		t.Fatalf("computed f4 = %v, want 16", got)
	}
}

func TestRequired(t *testing.T) {
	model := stubModel{
		name: "test.form.req",
		fields: []orm.FieldDefinition{
			{Name: "f1", Type: orm.Char, Required: true},
		},
	}
	f := mustNew(t, model)
	if err := f.Validate(); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
	if err := f.Set("f1", "x"); err != nil {
		t.Fatal(err)
	}
	if err := f.Validate(); err != nil {
		t.Fatalf("unexpected error after setting f1: %v", err)
	}
}

func TestOnchange(t *testing.T) {
	model := stubModel{
		name: "test.form.b",
		fields: []orm.FieldDefinition{
			{Name: "f1", Type: orm.Char},
			{Name: "f2", Type: orm.Char},
		},
	}
	orm.RegisterOnchange(model.name, "f1", func(_ context.Context, values map[string]interface{}, _ string) (orm.OnchangeResult, error) {
		return orm.OnchangeResult{Value: map[string]interface{}{"f2": "hello " + orm.AsString(values["f1"])}}, nil
	})

	f := mustNew(t, model)
	if err := f.Set("f1", "A"); err != nil {
		t.Fatal(err)
	}
	if got := f.Get("f2"); got != "hello A" {
		t.Fatalf("onchange f2 = %v, want %q", got, "hello A")
	}
}

func TestReadonly(t *testing.T) {
	model := stubModel{
		name: "test.form.ro",
		fields: []orm.FieldDefinition{
			{Name: "f1", Type: orm.Char, Readonly: true},
			{Name: "f2", Type: orm.Integer, Compute: "c", ComputeStore: false},
			{Name: "f3", Type: orm.Integer, Compute: "s", ComputeStore: true},
			{Name: "f4", Type: orm.Integer},
		},
	}
	f := mustNew(t, model)

	if err := f.Set("f1", "x"); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("f1: expected read-only error, got %v", err)
	}
	if err := f.Set("f2", 1); err == nil || !strings.Contains(err.Error(), "computed") {
		t.Fatalf("f2: expected computed error, got %v", err)
	}
	if err := f.Set("f3", 1); err == nil || !strings.Contains(err.Error(), "computed") {
		t.Fatalf("f3: expected computed error, got %v", err)
	}
	if err := f.Set("f4", 9); err != nil {
		t.Fatalf("f4 should be writable: %v", err)
	}
}

func TestWritableValues(t *testing.T) {
	model := stubModel{
		name: "test.form.w",
		fields: []orm.FieldDefinition{
			{Name: "ok", Type: orm.Integer},
			{Name: "virt", Type: orm.Integer, Compute: "c", ComputeStore: false},
			{Name: "stored", Type: orm.Integer, Compute: "s", ComputeStore: true},
			{Name: "rel", Type: orm.Char, Related: "x.name"},
			{Name: "m2m", Type: orm.Many2Many, Relation: "x"},
			{Name: "o2m", Type: orm.One2Many, Relation: "x"},
		},
	}
	f := mustNew(t, model)
	f.PutDraft("ok", 7)
	f.PutDraft("virt", 1)
	f.PutDraft("stored", 2)
	f.PutDraft("rel", "r")
	f.PutDraft("m2m", []interface{}{})
	f.PutDraft("o2m", []interface{}{})

	w := f.WritableValues()
	if len(w) != 1 || w["ok"] != 7 {
		t.Fatalf("writable values = %v, want {ok:7}", w)
	}
}
