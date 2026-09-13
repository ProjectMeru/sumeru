package orm_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/orm"
)

func TestRejectCoreUserSecurityWrites_nonAdmin(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 2)
	err := orm.RejectCoreUserSecurityWrites(ctx, 2, map[string]interface{}{
		"totp_secret": "SECRET",
	})
	if err == nil {
		t.Fatal("expected error for totp_secret write by non-admin")
	}
	if !strings.Contains(err.Error(), "system administrator") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRejectCoreUserSecurityWrites_allowsProfileFields(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 2)
	if err := orm.RejectCoreUserSecurityWrites(ctx, 2, map[string]interface{}{
		"name":  "Alice",
		"email": "a@example.com",
	}); err != nil {
		t.Fatalf("profile fields should be allowed: %v", err)
	}
}

func TestRejectCoreUserSecurityWrites_bypass(t *testing.T) {
	ctx := orm.ContextWithBypass(context.Background(), true)
	ctx = orm.ContextWithUID(ctx, 2)
	if err := orm.RejectCoreUserSecurityWrites(ctx, 2, map[string]interface{}{
		"active": false,
	}); err != nil {
		t.Fatalf("bypass should allow security fields: %v", err)
	}
}
