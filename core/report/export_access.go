package report

import (
	"context"

	"sumeru/core/orm"
)

// checkExportFieldAccess mirrors web requireExportEntitlement field checks for report entrypoints.
func checkExportFieldAccess(ctx context.Context, model string, fields []string) error {
	uid := orm.SecurityUID(ctx)
	if uid <= 0 || orm.SecurityBypass(ctx) {
		return nil
	}
	return orm.CheckFieldReadAccess(ctx, uid, model, fields)
}
