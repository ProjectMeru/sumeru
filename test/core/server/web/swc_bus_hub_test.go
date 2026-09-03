package web_test

import (
	"encoding/json"
	"testing"

	"sumeru/core/server/web"
)

func TestSwcBusHubBroadcastFiltersByActor(t *testing.T) {
	h := web.NewBusHubForTest()
	userA := web.NewSwcBusClientForTest(1, 1)
	userB := web.NewSwcBusClientForTest(2, 1)
	h.Register(userA)
	h.Register(userB)

	msg := []byte(`{"channel":"record.updated","payload":{"model":"crm.lead","id":1}}`)
	h.Broadcast(1, msg)

	select {
	case got := <-userA.Recv():
		if string(got) != string(msg) {
			t.Fatalf("user A: got %q", got)
		}
	default:
		t.Fatal("expected message for user A")
	}
	select {
	case <-userB.Recv():
		t.Fatal("user B should not receive actor-scoped message")
	default:
	}
}

func TestSwcBusHubQueueMessageShape(t *testing.T) {
	h := web.NewBusHubForTest()
	client := web.NewSwcBusClientForTest(5, 1)
	h.Register(client)

	raw, _ := json.Marshal(map[string]interface{}{
		"name":    "record.updated",
		"payload": map[string]interface{}{"model": "core.partner", "id": 3},
		"actor":   5,
	})
	var envelope map[string]interface{}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	name := envelope["name"].(string)
	inner := envelope["payload"].(map[string]interface{})
	out, _ := json.Marshal(map[string]interface{}{"channel": name, "payload": inner})
	h.Broadcast(5, out)

	select {
	case got := <-client.Recv():
		var parsed map[string]interface{}
		if err := json.Unmarshal(got, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed["channel"] != "record.updated" {
			t.Fatalf("channel: %v", parsed["channel"])
		}
	default:
		t.Fatal("expected bus message")
	}
}
