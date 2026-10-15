package report_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
	"sumeru/core/report"
)

func TestBulkJobActorContextRequiresPositiveUserID(t *testing.T) {
	ctx := report.BulkJobActorContextForTest(map[string]interface{}{"user_id": 0})
	if orm.UIDFromContext(ctx) != 0 {
		t.Fatalf("uid %d", orm.UIDFromContext(ctx))
	}
	if orm.SecurityBypass(ctx) {
		t.Fatal("must not carry bypass")
	}
	_ = context.Background()
}
