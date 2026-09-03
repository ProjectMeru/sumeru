package automation

import (
	"context"

	"sumeru/core/event"
)

func ExecuteServerActionForTest(ctx context.Context, row map[string]interface{}, ev event.Event) error {
	return executeServerAction(ctx, row, ev)
}

func RunServerActionsForEventForTest(ctx context.Context, ev event.Event) error {
	return runServerActionsForEvent(ctx, ev)
}
