package orm

import (
	"context"
	"fmt"
)

// SetUserActiveCompany updates core.user.company_id through the ORM write pipeline.
func SetUserActiveCompany(ctx context.Context, userID, companyID int) error {
	if userID <= 0 || companyID <= 0 {
		return fmt.Errorf("invalid user or company")
	}
	return UpdateRecordByID(ctx, "core.user", userID, map[string]interface{}{
		"company_id": companyID,
	})
}
