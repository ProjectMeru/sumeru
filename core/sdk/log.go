package sdk

import (
	"context"

	"sumeru/core/orm"
)

// AppendAudit writes one security audit row.
func AppendAudit(ctx context.Context, action, model string, resID int64, before, after map[string]interface{}, detail string) {
	orm.AppendAudit(ctx, action, model, resID, before, after, detail)
}

// AppendAppLog writes one application log row.
func AppendAppLog(ctx context.Context, moduleName, action, detail string) error {
	return orm.AppendAppLog(ctx, moduleName, action, detail)
}
