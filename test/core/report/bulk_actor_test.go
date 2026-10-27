package report_test

import (
	"testing"

	"sumeru/core/orm"
	"sumeru/core/report"
)

func TestBulkJobActorContextDropsBypass(t *testing.T) {
	ctx := report.BulkJobActorContextForTest(map[string]interface{}{"user_id": 7})
	if orm.SecurityBypass(ctx) {
		t.Fatal("queued job must not keep the scheduler bypass")
	}
	if orm.UIDFromContext(ctx) != 7 {
		t.Fatalf("uid %d", orm.UIDFromContext(ctx))
	}
}
