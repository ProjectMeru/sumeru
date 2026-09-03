package orm_test

import (
	"context"
	"testing"
	"sumeru/core/orm"
)


func TestCompanyIDFromContext(t *testing.T) {
	ctx := context.Background()
	if orm.CompanyIDFromContext(ctx) != 0 {
		t.Fatal("expected 0 for empty context")
	}
	ctx = orm.ContextWithCompanyID(ctx, 42)
	if got := orm.CompanyIDFromContext(ctx); got != 42 {
		t.Fatalf("company_id = %d, want 42", got)
	}
	// Child context with uid preserves company from parent.
	ctx = orm.ContextWithUID(ctx, 7)
	if got := orm.CompanyIDFromContext(ctx); got != 42 {
		t.Fatalf("company_id on child context = %d, want 42", got)
	}
}
