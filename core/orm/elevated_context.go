package orm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/applog"
)

type elevatedReasonKey struct{}

// ponytail: long module installs rely on this ceiling; increase if CI install times out.
const elevatedMaxDuration = 15 * time.Minute

func elevationActive(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	reason, ok := ctx.Value(elevatedReasonKey{}).(string)
	return ok && strings.TrimSpace(reason) != ""
}

func contextWithElevationMarker(ctx context.Context, reason string) context.Context {
	return context.WithValue(ContextWithBypass(ctx, true), elevatedReasonKey{}, reason)
}

// AuditedBypass returns an elevated context for internal kernel paths. Prefer WithElevated for bounded work units.
func AuditedBypass(ctx context.Context, reason string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(reason) == "" {
		return ctx
	}
	return ContextWithBypass(ctx, true)
}

// WithElevated runs fn with security bypass, a time box, and sys.audit rows (internal/cron/install paths only).
func WithElevated(ctx context.Context, reason string, fn func(context.Context) error) error {
	if fn == nil {
		return nil
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("orm.WithElevated: reason required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if elevationActive(ctx) {
		return fn(AuditedBypass(ctx, reason))
	}

	AppendAudit(ctx, "security_elevate", "sys.security", 0, nil, nil, reason)
	timeoutCtx, cancel := context.WithTimeout(ctx, elevatedMaxDuration)
	defer cancel()
	work := contextWithElevationMarker(timeoutCtx, reason)

	err := fn(work)
	if err != nil {
		failDetail := reason + ": " + fmt.Sprint(applog.ScrubValue("error", err.Error()))
		AppendAudit(ctx, "security_elevate_fail", "sys.security", 0, nil, nil, failDetail)
		applog.WarnMsg(work, "orm", "elevated", "elevated context failed",
			err, map[string]interface{}{"reason": reason})
		return err
	}
	AppendAudit(ctx, "security_elevate_done", "sys.security", 0, nil, nil, reason)
	return nil
}
