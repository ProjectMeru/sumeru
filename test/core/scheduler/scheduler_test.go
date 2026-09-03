package scheduler_test

import (
	"context"
	"testing"
	"time"

	"sumeru/core/scheduler"
)

func TestStartStop_idempotent(t *testing.T) {
	ctx := context.Background()
	scheduler.Start(ctx, 50*time.Millisecond)
	scheduler.Start(ctx, 50*time.Millisecond) // second start ignored
	time.Sleep(80 * time.Millisecond)         // let loop tick once with nil DB
	scheduler.Stop()
	scheduler.Stop() // second stop safe
}
