package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"

	"sumeru/core/scheduler"
)

func TestRegisterCronHandler(t *testing.T) {
	scheduler.ClearCronHandlers()
	var called int32
	scheduler.RegisterCronHandler("test_job", func(_ context.Context, payload map[string]interface{}) error {
		if payload["code"] != "test_job" {
			t.Fatalf("unexpected payload: %v", payload)
		}
		atomic.StoreInt32(&called, 1)
		return nil
	})
	scheduler.ExecuteCronForTest(context.Background(), scheduler.CronRunInput{
		ID: 1, Name: "Test", Code: "test_job",
	})
	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("cron handler was not called")
	}
	scheduler.ClearCronHandlers()
}
