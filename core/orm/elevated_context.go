package orm

import (
	"context"

	"sumeru/core/applog"
)

// AuditedBypass returns an elevated context for internal kernel paths. Prefer WithElevated for bounded work units.
func AuditedBypass(ctx context.Context, reason string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	applog.DebugMsg(ctx, "orm", "elevated", "elevated context entered",
		map[string]interface{}{
			"reason": reason,
			"uid":    SecurityUID(ctx),
		})
	return ContextWithBypass(ctx, true)
}

// WithElevated runs fn with security bypass and emits an audit event (internal/cron paths only).
func WithElevated(ctx context.Context, reason string, fn func(context.Context) error) error {
	if fn == nil {
		return nil
	}
	elevated := AuditedBypass(ctx, reason)
	err := fn(elevated)
	if err != nil {
		applog.WarnMsg(elevated, "orm", "elevated", "elevated context failed",
			err, map[string]interface{}{"reason": reason})
	}
	return err
}
