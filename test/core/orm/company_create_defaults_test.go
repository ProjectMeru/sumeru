package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestStampActiveCompanyOnCreate_setsFromContext(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.co.stamp", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "company_id", Type: orm.Many2One, Relation: "core.company"},
	})
	ctx := orm.ContextWithCompanyID(orm.ContextWithUID(context.Background(), 7), 3)
	prepared := map[string]interface{}{"name": "Row"}
	orm.StampActiveCompanyOnCreateForTest(ctx, "test.co.stamp", prepared)
	if prepared["company_id"] != 3 {
		t.Fatalf("company_id = %v, want 3", prepared["company_id"])
	}
}

func TestStampActiveCompanyOnCreate_skipsWhenSet(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.co.stamp2", []orm.FieldDefinition{
		{Name: "company_id", Type: orm.Many2One, Relation: "core.company"},
	})
	ctx := orm.ContextWithCompanyID(orm.ContextWithUID(context.Background(), 7), 3)
	prepared := map[string]interface{}{"company_id": 9}
	orm.StampActiveCompanyOnCreateForTest(ctx, "test.co.stamp2", prepared)
	if prepared["company_id"] != 9 {
		t.Fatalf("company_id = %v, want 9 (unchanged)", prepared["company_id"])
	}
}
