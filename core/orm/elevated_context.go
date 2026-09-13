package orm

import (
	"context"

	"sumeru/core/applog"
)

// WithElevated runs fn with security bypass and emits an audit event (internal/cron paths only).
func WithElevated(ctx context.Context, reason string, fn func(context.Context) error) error {
	if fn == nil {
		return nil
	}
	elevated := ContextWithBypass(ctx, true)
	applog.InfoMsg(elevated, "orm", "elevated", "elevated context entered",
		map[string]interface{}{
			"reason": reason,
			"uid":    SecurityUID(ctx),
		})
	err := fn(elevated)
	if err != nil {
		applog.WarnMsg(elevated, "orm", "elevated", "elevated context failed",
			err, map[string]interface{}{"reason": reason})
	}
	return err
}
