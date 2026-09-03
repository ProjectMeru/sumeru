package queue_test

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sumeru/core/queue"
)

func TestPublishSubscribe(t *testing.T) {
	var received atomic.Int32
	var payload string
	done := make(chan struct{})
	queue.Subscribe("test.topic", func(ctx context.Context, msg queue.Message) error {
		received.Add(1)
		payload = string(msg.Payload)
		close(done)
		return nil
	})
	queue.Publish(context.Background(), "test.topic", map[string]string{"hello": "world"})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
	if received.Load() != 1 {
		t.Fatalf("received=%d", received.Load())
	}
	var decoded map[string]string
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil || decoded["hello"] != "world" {
		t.Fatalf("payload=%q err=%v", payload, err)
	}
}

func TestPublishMirrorHook(t *testing.T) {
	queue.ClearPublishMirrorHook()
	t.Cleanup(queue.ClearPublishMirrorHook)

	var mu sync.Mutex
	var mirroredTopic string
	var mirroredPayload []byte
	queue.SetPublishMirrorHook(func(ctx context.Context, topic string, payload []byte) {
		mu.Lock()
		mirroredTopic = topic
		mirroredPayload = append([]byte(nil), payload...)
		mu.Unlock()
	})
	queue.Publish(context.Background(), "mirror.topic", "payload")
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if mirroredTopic != "mirror.topic" || string(mirroredPayload) != `"payload"` {
		t.Fatalf("mirror topic=%q payload=%q", mirroredTopic, mirroredPayload)
	}
}

func TestPublishInvalidPayloadIgnored(t *testing.T) {
	queue.Publish(context.Background(), "bad", make(chan int))
}
