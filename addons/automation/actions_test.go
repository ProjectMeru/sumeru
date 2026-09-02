package automation

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/event"
)

func TestExecuteServerActionPublishPrefix(t *testing.T) {
	ctx := context.Background()
	var got string
	event.Subscribe("test.followup", func(_ context.Context, ev event.Event) error {
		got = ev.Name
		return nil
	})
	defer event.Clear()

	row := map[string]interface{}{
		"name": "Test Action",
		"code": "publish:test.followup",
	}
	ev := event.Event{Name: "record.created", Payload: map[string]interface{}{"model": "crm.lead", "id": 1}}
	if err := executeServerAction(ctx, row, ev); err != nil {
		t.Fatal(err)
	}
	if got != "test.followup" {
		t.Fatalf("expected test.followup, got %q", got)
	}
}

func TestExecuteServerActionModelFilter(t *testing.T) {
	ctx := context.Background()
	var got string
	event.Subscribe("test.filtered", func(_ context.Context, ev event.Event) error {
		got = ev.Name
		return nil
	})
	defer event.Clear()

	row := map[string]interface{}{
		"name":  "Filtered",
		"model": "sale.order",
		"code":  "publish:test.filtered",
	}
	ev := event.Event{Name: "record.updated", Payload: map[string]interface{}{"model": "crm.lead", "id": 1}}
	if err := executeServerAction(ctx, row, ev); err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("expected no publish, got %q", got)
	}
}

func TestExecuteServerActionWebhook(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	row := map[string]interface{}{
		"name": "Webhook",
		"code": "webhook:" + srv.URL,
	}
	ev := event.Event{Name: "record.created", Payload: map[string]interface{}{"model": "crm.lead", "id": 1}}
	if err := executeServerAction(context.Background(), row, ev); err != nil {
		t.Fatal(err)
	}
	if gotBody == "" || !strings.Contains(gotBody, "record.created") {
		t.Fatalf("expected webhook body with event, got %q", gotBody)
	}
}
