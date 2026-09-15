package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestPortalAllowsModel(t *testing.T) {
	if !orm.PortalAllowsModel("core.partner", "read") {
		t.Fatal("core.partner should be allowlisted")
	}
	if orm.PortalAllowsModel("core.user", "read") {
		t.Fatal("core.user must not be portal allowlisted")
	}
}

func TestEnforcePortalRPCModel(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 99)
	if err := orm.EnforcePortalRPCModel(ctx, "core.partner"); err != nil {
		t.Fatalf("non-portal user: %v", err)
	}
}
