package im

import (
	"context"
	"fmt"

	"sumeru/core/sdk"
)

// CompanySettingsFormHref returns a workspace URL to edit company IM settings when the user may read core.company.
func CompanySettingsFormHref(ctx context.Context, companyID int) string {
	if 0 >= companyID {
		return ""
	}
	uid := sdk.SecurityUID(ctx)
	if 0 >= uid {
		return ""
	}
	if err := sdk.CheckModelAccess(ctx, uid, "core.company", "read"); err != nil {
		return ""
	}
	return fmt.Sprintf("/web?model=core.company&view_type=form&id=%d", companyID)
}
