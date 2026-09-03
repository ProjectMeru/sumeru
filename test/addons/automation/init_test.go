package automation_test

import (
	"sumeru/addons/automation"
	"context"
	"testing"
	"sumeru/core/event"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)



func TestShouldMaterializeServerActionRequiresAutomationModule(t *testing.T) {
	_ = modelreg.ActivateAll([]string{"base", "mail", "automation"})
	installed := map[string]struct{}{"base": {}, "mail": {}}
	if orm.ShouldMaterializeModel("sys.server.action", installed) {
		t.Fatal("sys.server.action should not materialize without automation module")
	}
	installed["automation"] = struct{}{}
	if !orm.ShouldMaterializeModel("sys.server.action", installed) {
		t.Fatal("sys.server.action should materialize when automation is installed")
	}
}

func TestRunServerActionsForEventNilDB(t *testing.T) {
	if orm.DB != nil {
		t.Skip("requires orm.DB unset")
	}
	ev := event.Event{Name: "record.updated", Payload: map[string]interface{}{"model": "core.partner", "id": 1}}
	if err := automation.RunServerActionsForEventForTest(context.Background(), ev); err != nil {
		t.Fatalf("expected nil when DB unavailable, got %v", err)
	}
}
