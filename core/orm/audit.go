package orm

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"sumeru/core/applog"
	"sumeru/core/errcode"
)

func skipAuditModel(model string) bool {
	switch model {
	case "sys.audit", "sys.session", "app.log", "core.user.log", "mail.message", "sys.outbox.event":
		return true
	default:
		return false
	}
}

func auditValues(ctx context.Context, action, model string, resID int64, before, after map[string]interface{}, detail string) map[string]interface{} {
	uid := SecurityUID(ctx)
	var beforeJSON, afterJSON string
	if before != nil {
		beforeJSON = marshalAuditJSON(ctx, "before_json", model, resID, before)
	}
	if after != nil {
		afterJSON = marshalAuditJSON(ctx, "after_json", model, resID, after)
	}
	vals := map[string]interface{}{
		"action":      action,
		"model":       strings.TrimSpace(model),
		"res_id":      resID,
		"before_json": beforeJSON,
		"after_json":  afterJSON,
		"detail":      detail,
		"create_date": time.Now().UTC().Format(time.RFC3339),
	}
	if uid > 0 {
		vals["user_id"] = uid
	}
	return vals
}

func marshalAuditJSON(ctx context.Context, field, model string, resID int64, values map[string]interface{}) string {
	b, err := json.Marshal(scrubAuditMap(values))
	if err != nil {
		applog.WarnCode(ctx, errcode.InternalError, "Audit "+field+" marshal failed", applog.Event{
			Component: "orm",
			Operation: "audit",
			Status:    "partial",
			Context:   map[string]interface{}{"resource": model, "resource_id": resID},
			Err:       err,
		})
		return ""
	}
	return string(b)
}

// AppendAudit writes an immutable audit row (best-effort; never fails the caller).
func AppendAudit(ctx context.Context, action, model string, resID int64, before, after map[string]interface{}, detail string) {
	AppendAuditTx(ctx, nil, action, model, resID, before, after, detail)
}

// AppendAuditTx writes an audit row on tx when non-nil; otherwise uses the pool.
func AppendAuditTx(ctx context.Context, tx TxWrapper, action, model string, resID int64, before, after map[string]interface{}, detail string) {
	if strings.TrimSpace(action) == "" || skipAuditModel(model) {
		return
	}
	vals := auditValues(ctx, action, model, resID, before, after, detail)
	_ = insertSideEffectRow(ctx, tx, "sys.audit", vals)
}

func scrubAuditMap(m map[string]interface{}) map[string]interface{} {
	return applog.ScrubMap(m)
}

// LogAccessDeny records a permission denial in sys.audit.
func LogAccessDeny(ctx context.Context, model, op, detail string) {
	AppendAudit(ctx, "access_deny", model, 0, nil, nil, op+": "+detail)
}
