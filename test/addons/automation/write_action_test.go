package automation_test

import (
	"context"
	"testing"

	"sumeru/addons/automation"
	"sumeru/core/event"
)

func TestExecuteServerActionWritePrefix_noPayload(t *testing.T) {
	row := map[string]interface{}{
		"name": "Write",
		"code": `write:{"state":"done"}`,
	}
	ev := event.Event{Name: "record.updated", Payload: map[string]interface{}{}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteServerActionWritePrefix_invalidJSON(t *testing.T) {
	row := map[string]interface{}{
		"name": "Write",
		"code": `write:{bad json}`,
	}
	ev := event.Event{
		Name:    "record.updated",
		Payload: map[string]interface{}{"model": "test.model", "id": int64(1)},
	}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteServerActionEmptyCode(t *testing.T) {
	row := map[string]interface{}{"name": "Noop", "code": ""}
	ev := event.Event{Name: "x", Payload: map[string]interface{}{"model": "m", "id": int64(1)}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteServerActionUnknownPrefix(t *testing.T) {
	row := map[string]interface{}{"name": "Unknown", "code": "noop:thing"}
	ev := event.Event{Name: "x", Payload: map[string]interface{}{"model": "m", "id": int64(1)}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteServerActionTriggerDomain_noRecord(t *testing.T) {
	row := map[string]interface{}{
		"name":           "Domain",
		"code":           "publish:test.event",
		"trigger_domain": `[["active","=",true]]`,
	}
	ev := event.Event{Name: "record.created", Payload: map[string]interface{}{"model": "crm.lead"}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteServerActionPublishEmptyName(t *testing.T) {
	row := map[string]interface{}{"name": "Pub", "code": "publish:  "}
	ev := event.Event{Name: "x", Payload: map[string]interface{}{"model": "m", "id": int64(1)}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteServerActionWebhookEmptyURL(t *testing.T) {
	row := map[string]interface{}{"name": "Hook", "code": "webhook:  "}
	ev := event.Event{Name: "x", Payload: map[string]interface{}{"model": "m", "id": int64(1)}}
	if err := automation.ExecuteServerActionForTest(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
}
