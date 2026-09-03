package scheduler

import "context"

// ExecuteCronForTest runs one cron tick handler (tests).
func ExecuteCronForTest(ctx context.Context, in CronRunInput) {
	executeCron(ctx, in)
}
