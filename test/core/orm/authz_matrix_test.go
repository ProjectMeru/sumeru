package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestAuthzMatrix_anonymousDenied(t *testing.T) {
	ctx := context.Background()
	err := orm.CheckModelAccess(ctx, 0, "core.user", "read")
	if err == nil {
		t.Fatal("anonymous must be denied")
	}
	if !orm.IsAccessDenied(err) {
		t.Fatalf("want AccessDenied, got %v", err)
	}
}

func TestAuthzMatrix_bypassAllows(t *testing.T) {
	ctx := orm.ContextWithBypass(context.Background(), true)
	if err := orm.CheckModelAccess(ctx, 0, "core.user", "write"); err != nil {
		t.Fatalf("bypass should allow: %v", err)
	}
}

func TestAuthzMatrix_superuserAllows(t *testing.T) {
	ctx := context.Background()
	if err := orm.CheckModelAccess(ctx, 1, "core.user", "write"); err != nil {
		t.Fatalf("uid=1 should allow: %v", err)
	}
}

func TestAuthzMatrix_unknownModel(t *testing.T) {
	err := orm.CheckModelAccess(context.Background(), 2, "no.such.model", "read")
	if err == nil {
		t.Fatal("expected unknown model error")
	}
}

func TestAuthzMatrix_passwordPrepareDenied(t *testing.T) {
	_, err := orm.PrepareValues(stubUserModel{}, map[string]interface{}{"password": "x"}, orm.WriteOpWrite, orm.PrepareOptions{})
	if err == nil {
		t.Fatal("password write must be rejected")
	}
}
