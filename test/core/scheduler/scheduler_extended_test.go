package scheduler_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"sumeru/core/event"
	"sumeru/core/scheduler"
)

func TestExecuteCronForTest_publishesEvents(t *testing.T) {
	event.Clear()
	t.Cleanup(event.Clear)
	scheduler.ClearCronHandlers()
	t.Cleanup(scheduler.ClearCronHandlers)

	var tickCount int32
	event.Subscribe("cron.tick", func(_ context.Context, _ event.Event) error {
		atomic.AddInt32(&tickCount, 1)
		return nil
	})
	var customCount int32
	event.Subscribe("custom.event", func(_ context.Context, _ event.Event) error {
		atomic.AddInt32(&customCount, 1)
		return nil
	})

	scheduler.ExecuteCronForTest(context.Background(), scheduler.CronRunInput{
		ID: 1, Name: "Daily", EventName: "custom.event", Code: "missing_handler",
	})
	if atomic.LoadInt32(&tickCount) != 1 || atomic.LoadInt32(&customCount) != 1 {
		t.Fatalf("events: tick=%d custom=%d", tickCount, customCount)
	}
}

func TestExecuteCronForTest_handlerError(t *testing.T) {
	scheduler.ClearCronHandlers()
	t.Cleanup(scheduler.ClearCronHandlers)
	scheduler.RegisterCronHandler("fail_job", func(_ context.Context, _ map[string]interface{}) error {
		return errors.New("boom")
	})
	scheduler.ExecuteCronForTest(context.Background(), scheduler.CronRunInput{
		ID: 2, Name: "Fail", Code: "fail_job",
	})
}

func TestRegisterCronHandler_ignoresBlank(t *testing.T) {
	scheduler.ClearCronHandlers()
	t.Cleanup(scheduler.ClearCronHandlers)
	scheduler.RegisterCronHandler("", func(context.Context, map[string]interface{}) error { return nil })
	scheduler.RegisterCronHandler("x", nil)
	scheduler.ExecuteCronForTest(context.Background(), scheduler.CronRunInput{Code: ""})
}
