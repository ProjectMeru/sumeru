package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestCheckFieldWriteAccess_bypassAllows(t *testing.T) {
	ctx := orm.ContextWithBypass(context.Background(), true)
	err := orm.CheckFieldWriteAccess(ctx, 2, "core.user", map[string]interface{}{
		"password": "blocked-for-normal-users",
	})
	if err != nil {
		t.Fatalf("bypass should allow field write check: %v", err)
	}
}

func TestCheckFieldWriteAccess_superuserAllows(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 1)
	err := orm.CheckFieldWriteAccess(ctx, 1, "core.user", map[string]interface{}{
		"name": "Admin",
	})
	if err != nil {
		t.Fatalf("superuser should pass field write check: %v", err)
	}
}
