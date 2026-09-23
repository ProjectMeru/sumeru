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

func TestAuthzMatrix_smuggledBypassRejected(t *testing.T) {
	ctx := orm.ContextWithUID(orm.ContextWithBypass(context.Background(), true), 42)
	err := orm.RejectSmuggledUserBypass(ctx)
	if err == nil || !orm.IsSmuggledBypass(err) {
		t.Fatalf("want SmuggledBypassError, got %v", err)
	}
}

func TestAuthzMatrix_smuggledBypassAllowedInElevation(t *testing.T) {
	parent := orm.ContextWithUID(context.Background(), 42)
	err := orm.WithElevated(parent, "test.authz", func(elevated context.Context) error {
		if err := orm.RejectSmuggledUserBypass(elevated); err != nil {
			t.Fatalf("elevated boundary: %v", err)
		}
		if !orm.BypassFromContext(elevated) {
			t.Fatal("expected bypass inside WithElevated")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithElevated: %v", err)
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
