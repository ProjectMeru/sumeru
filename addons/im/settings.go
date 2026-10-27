package im

import (
	"context"

	"sumeru/core/sdk"
)

// IMEnabled reports whether internal chat is enabled for the active company (default true).
func IMEnabled(ctx context.Context) bool {
	if !sdk.DatabaseOpen() {
		return true
	}
	companyID := sdk.CompanyIDFromContext(ctx)
	if 0 >= companyID {
		companyID = sdk.ActiveCompanyIDForUser(ctx, sdk.SecurityUID(ctx))
	}
	rec, err := companyRow(ctx, int(companyID))
	if err != nil || 0 == len(rec) {
		return true
	}
	if _, ok := rec["im_enabled"]; !ok {
		return true
	}
	return sdk.AsBool(rec["im_enabled"])
}

func companyRow(ctx context.Context, companyID int) (map[string]interface{}, error) {
	if companyID > 0 {
		return sdk.SearchOne(ctx, "core.company", map[string]interface{}{"id": companyID})
	}
	rows, err := sdk.SearchPage(ctx, "core.company", nil, 1, 0, "id ASC")
	if err != nil || 0 == len(rows) {
		return nil, err
	}
	return rows[0], nil
}
