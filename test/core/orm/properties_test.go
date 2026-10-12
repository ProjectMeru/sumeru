package orm_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"sumeru/core/orm"
)

func TestPropertiesRejectsUnsafeKeysAndTypes(t *testing.T) {
	const model = "test.properties.record"
	orm.RegisterStubModelForTest(t, model, []orm.FieldDefinition{
		{Name: "props", Type: orm.Properties},
	})
	orm.PropertyDefLoader = func(modelName, field string) (map[string]orm.PropertyDef, error) {
		return map[string]orm.PropertyDef{
			"label":  {Key: "label", ValueType: "char", Required: true},
			"qty":    {Key: "qty", ValueType: "integer"},
			"choice": {Key: "choice", ValueType: "selection", Selection: "a:A,b:B"},
			"owner":  {Key: "owner", ValueType: "many2one"},
		}, nil
	}
	t.Cleanup(func() { orm.PropertyDefLoader = nil })

	if _, err := orm.CoerceFieldValueForTest(orm.FieldDefinition{Name: "props", Type: orm.Properties}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := orm.CoerceFieldValueForTest(orm.FieldDefinition{Name: "props", Type: orm.Properties}, "{}"); err != nil {
		t.Fatal(err)
	}
	if _, err := orm.CoerceFieldValueForTest(orm.FieldDefinition{Name: "props", Type: orm.Properties}, `{"__proto__":1}`); err == nil {
		t.Fatal("expected forbidden key to fail")
	}

	prepared := map[string]interface{}{"props": map[string]interface{}{"label": "ok", "qty": 2, "choice": "a", "owner": 3}}
	if err := orm.MergePropertiesWrite(context.Background(), model, nil, prepared); err != nil {
		t.Fatal(err)
	}
	kept := map[string]interface{}{"props": map[string]interface{}{"label": "ok"}}
	before := map[string]interface{}{"props": `{"retired":"stay"}`}
	if err := orm.MergePropertiesWrite(context.Background(), model, before, kept); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fmt.Sprint(kept["props"]), "retired") {
		t.Fatalf("removed definition key was dropped: %v", kept["props"])
	}
	bad := map[string]interface{}{"props": map[string]interface{}{"label": "ok", "extra": 1}}
	if err := orm.MergePropertiesWrite(context.Background(), model, nil, bad); err == nil {
		t.Fatal("expected unknown key to fail")
	}
	missing := map[string]interface{}{"props": map[string]interface{}{"qty": 1}}
	if err := orm.MergePropertiesWrite(context.Background(), model, nil, missing); err == nil {
		t.Fatal("expected required key to fail")
	}
	choice := map[string]interface{}{"props": map[string]interface{}{"label": "ok", "choice": "nope"}}
	if err := orm.MergePropertiesWrite(context.Background(), model, nil, choice); err == nil {
		t.Fatal("expected selection to fail")
	}
	owner := map[string]interface{}{"props": map[string]interface{}{"label": "ok", "owner": 0}}
	if err := orm.MergePropertiesWrite(context.Background(), model, nil, owner); err == nil {
		t.Fatal("expected many2one to fail")
	}
}

func TestPropertiesDomainKeyIsLiteral(t *testing.T) {
	const model = "test.properties.domain"
	orm.RegisterStubModelForTest(t, model, []orm.FieldDefinition{
		{Name: "props", Type: orm.Properties, Column: "props"},
	})
	orm.PropertyDefLoader = func(modelName, field string) (map[string]orm.PropertyDef, error) {
		return map[string]orm.PropertyDef{
			"label": {Key: "label", ValueType: "char"},
			"qty":   {Key: "qty", ValueType: "integer"},
		}, nil
	}
	t.Cleanup(func() { orm.PropertyDefLoader = nil })

	where, args, err := orm.BuildSearchWhereClauseForTest(model, [][]interface{}{{"props.label", "ilike", "red"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(where, "->>'label'") || strings.Contains(where, "red") {
		t.Fatalf("where %q args %v", where, args)
	}
	if _, _, err := orm.BuildSearchWhereClauseForTest(model, [][]interface{}{{"props.missing", "=", "x"}}); err == nil {
		t.Fatal("expected undefined key to fail")
	}
	if _, _, err := orm.BuildSearchWhereClauseForTest(model, [][]interface{}{{"props.qty", "ilike", "1"}}); err == nil {
		t.Fatal("expected ilike on integer to fail")
	}
	if _, _, err := orm.BuildSearchWhereClauseForTest(model, [][]interface{}{{"props.qty';drop", "=", "1"}}); err == nil {
		t.Fatal("expected unsafe key to fail")
	}
}
