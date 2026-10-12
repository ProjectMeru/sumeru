package automation_test

import (
	"context"
	"testing"

	"sumeru/addons/automation"
	"sumeru/core/event"
	"sumeru/core/orm"
	"sumeru/core/security"
)

func TestExecuteServerActionDepthCap(t *testing.T) {
	ctx := automation.WithAutomationDepthForTest(context.Background(), 5)
	row := map[string]interface{}{"name": "Deep", "code": "publish:test.depth"}
	ev := event.Event{Name: "record.created", Payload: map[string]interface{}{"model": "crm.lead", "id": int64(1)}}
	if err := automation.ExecuteServerActionForTest(ctx, row, ev); err == nil {
		t.Fatal("expected depth cap to stop publish")
	}
}

func TestExecuteServerActionInvalidDomainFailsClosed(t *testing.T) {
	var got string
	event.Subscribe("test.bad.domain", func(_ context.Context, ev event.Event) error {
		got = ev.Name
		return nil
	})
	defer event.Clear()
	row := map[string]interface{}{
		"name":           "Bad",
		"code":           "publish:test.bad.domain",
		"trigger_domain": "{",
	}
	ev := event.Event{Name: "record.created", Payload: map[string]interface{}{"model": "crm.lead", "id": int64(1)}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("invalid domain published %q", got)
	}
}

func TestExecuteServerActionRejectsSensitiveWrite(t *testing.T) {
	const model = "test.server.action.write"
	orm.RegisterStubModelForTest(t, model, []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "password", Type: orm.Char},
	})
	prev := security.FieldRegistry[model]
	security.FieldRegistry[model] = map[string]security.FieldPolicy{
		"password": {WriteDenyDirect: true},
	}
	t.Cleanup(func() {
		if prev == nil {
			delete(security.FieldRegistry, model)
			return
		}
		security.FieldRegistry[model] = prev
	})
	row := map[string]interface{}{"name": "Write", "code": `write:{"password":"secret","name":"ok"}`}
	ev := event.Event{Name: "record.updated", Payload: map[string]interface{}{"model": model, "id": int64(1)}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err == nil {
		t.Fatal("expected sensitive write to be rejected")
	}
	unknown := map[string]interface{}{"name": "Write", "code": `write:{"missing":1}`}
	if err := automation.ExecuteServerActionForTest(context.Background(), unknown, ev); err == nil {
		t.Fatal("expected unknown field to be rejected")
	}
}

func TestExecuteServerActionActivityRequiresType(t *testing.T) {
	row := map[string]interface{}{"name": "Act", "code": "activity:mail.activity_type_todo|Call"}
	ev := event.Event{Name: "record.created", Payload: map[string]interface{}{"model": "crm.lead", "id": int64(1)}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err == nil {
		t.Fatal("expected missing activity type to fail")
	}
}
