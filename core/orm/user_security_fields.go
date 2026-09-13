package orm

import (
	"context"
	"fmt"
)

// coreUserSecurityWriteFields may only be written by system administrators (or bypass/superuser).
var coreUserSecurityWriteFields = map[string]bool{
	"totp_secret":  true,
	"totp_enabled": true,
	"active":       true,
	"user_type":    true,
	"company_id":   true,
}

// RejectCoreUserSecurityWrites blocks mass-assignment of security-sensitive core.user columns.
func RejectCoreUserSecurityWrites(ctx context.Context, uid int, values map[string]interface{}) error {
	if len(values) == 0 {
		return nil
	}
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil
	}
	if UserHasGroupXML(ctx, uid, "base.group_system") {
		return nil
	}
	for k := range values {
		if k == "id" {
			continue
		}
		if coreUserSecurityWriteFields[k] {
			return fmt.Errorf("field %q on core.user requires system administrator", k)
		}
	}
	return nil
}
