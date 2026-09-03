package event_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"sumeru/core/event"
)

func TestPublish_callsHandlersInOrder(t *testing.T) {
	event.Clear()
	t.Cleanup(event.Clear)

	var order []int
	event.Subscribe("t.order", func(ctx context.Context, ev event.Event) error {
		order = append(order, 1)
		return nil
	})
	event.Subscribe("t.order", func(ctx context.Context, ev event.Event) error {
		order = append(order, 2)
		return nil
	})
	errs := event.Publish(context.Background(), event.Event{Name: "t.order"})
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Fatalf("order = %v", order)
	}
}

func TestPublish_collectsHandlerErrors(t *testing.T) {
	event.Clear()
	t.Cleanup(event.Clear)

	event.Subscribe("t.err", func(ctx context.Context, ev event.Event) error {
		return errors.New("boom")
	})
	event.Subscribe("t.err", func(ctx context.Context, ev event.Event) error {
		return nil
	})
	errs := event.Publish(context.Background(), event.Event{Name: "t.err"})
	if len(errs) != 1 || errs[0].Error() != "boom" {
		t.Fatalf("errs = %v", errs)
	}
}

func TestSubscribe_ignoresEmpty(t *testing.T) {
	event.Clear()
	t.Cleanup(event.Clear)
	var n atomic.Int32
	event.Subscribe("", func(ctx context.Context, ev event.Event) error {
		n.Add(1)
		return nil
	})
	event.Subscribe("t.nil", nil)
	_ = event.Publish(context.Background(), event.Event{Name: "t.nil"})
	if n.Load() != 0 {
		t.Fatal("empty/nil subscribe should not register")
	}
}
