package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"sumeru/core/applog"
	"sumeru/core/event"
)

func dispatchWebhook(ctx context.Context, url string, ev event.Event) error {
	body, err := json.Marshal(map[string]interface{}{
		"event":   ev.Name,
		"actor":   ev.Actor,
		"payload": ev.Payload,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		applog.Warn(ctx, applog.Event{
			Message:   "webhook server action failed",
			Component: "automation",
			Operation: "webhook",
			Status:    "failed",
			Context:   map[string]interface{}{"url": url, "status": resp.StatusCode},
		})
	}
	return nil
}
