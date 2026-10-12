package automation

import (
	"context"

	"sumeru/core/event"
)

func ValidateWebhookURLForTest(raw string) error {
	return validateWebhookURL(raw)
}

func ExecuteServerActionForTest(ctx context.Context, row map[string]interface{}, ev event.Event) error {
	return executeServerAction(ctx, row, ev)
}

func WithAutomationDepthForTest(ctx context.Context, n int) context.Context {
	return withDepth(ctx, n)
}

func RunServerActionsForEventForTest(ctx context.Context, ev event.Event) error {
	return runServerActionsForEvent(ctx, ev)
}
