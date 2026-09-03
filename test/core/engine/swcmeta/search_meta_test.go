package swcmeta_test

import (
	"context"
	"testing"
	"sumeru/core/engine/swcmeta"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)


type stubFilterModel struct{}

func (stubFilterModel) ModelName() string { return "demo.filter.model" }

func (stubFilterModel) Fields() []orm.FieldDefinition {
	return []orm.FieldDefinition{
		{Name: "name", Type: orm.Char, String: "Name"},
		{Name: "priority", Type: orm.Integer, String: "Priority"},
		{Name: "state", Type: orm.Selection, String: "Status", Selection: [][]string{{"draft", "Draft"}}},
		{Name: "line_ids", Type: orm.One2Many, String: "Lines"},
	}
}

func TestFilterableAndGroupByFields(t *testing.T) {
	orm.Registry["demo.filter.model"] = stubFilterModel{}
	t.Cleanup(func() { delete(orm.Registry, "demo.filter.model") })

	filterable := swcmeta.FilterableFields("demo.filter.model")
	if len(filterable) < 3 {
		t.Fatalf("expected filterable fields, got %d", len(filterable))
	}
	groupBy := swcmeta.GroupByFields("demo.filter.model")
	names := map[string]bool{}
	for _, f := range groupBy {
		names[f.Name] = true
	}
	if !names["state"] || names["line_ids"] {
		t.Fatalf("groupBy fields: %+v", groupBy)
	}
	_ = modelreg.MustRegister
	_ = context.Background()
}

func TestBuildSearchMetaIncludesModelFields(t *testing.T) {
	orm.Registry["demo.filter.model"] = stubFilterModel{}
	t.Cleanup(func() { delete(orm.Registry, "demo.filter.model") })

	meta := swcmeta.BuildSearchMeta(context.Background(), "demo.filter.model", nil)
	if meta == nil {
		t.Fatal("expected search meta")
	}
	if len(meta.FilterFields) == 0 || len(meta.GroupByFields) == 0 {
		t.Fatalf("meta: %+v", meta)
	}
}
