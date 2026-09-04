// Package queue provides an in-process pub/sub bus; optional external brokers can hook Publish.
package queue

import (
	"context"
	"encoding/json"
	"sync"

	"sumeru/core/applog"
	"sumeru/core/metrics"
)

// Message is a topic-tagged payload for async workers.
type Message struct {
	Topic   string
	Payload json.RawMessage
}

type handler func(ctx context.Context, msg Message) error

var (
	mu       sync.RWMutex
	handlers = map[string][]handler{}
)

// Subscribe registers fn for topic (multiple subscribers allowed).
func Subscribe(topic string, fn func(ctx context.Context, msg Message) error) {
	mu.Lock()
	defer mu.Unlock()
	handlers[topic] = append(handlers[topic], fn)
}

// Publish dispatches payload to subscribers asynchronously (in-process).
func Publish(ctx context.Context, topic string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	msg := Message{Topic: topic, Payload: data}

	mu.RLock()
	subs := append([]handler(nil), handlers[topic]...)
	mu.RUnlock()
	for _, fn := range subs {
		fn := fn
		go func() {
			if err := fn(ctx, msg); err != nil {
				metrics.Inc("sumeru_queue_handler_errors_total")
				applog.Warn(ctx, applog.Event{
					Message:   "queue handler failed",
					Component: "queue",
					Operation: "publish",
					Status:    "failed",
					Context:   map[string]interface{}{"topic": topic},
					Err:       err,
				})
			}
		}()
	}
	publishRedisMirror(ctx, topic, data)
}
