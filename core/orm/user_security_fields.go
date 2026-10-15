package orm

import (
	"context"
	"fmt"
)

// RejectCoreUserSecurityWrites blocks mass-assignment of security-sensitive core.user columns.
func RejectCoreUserSecurityWrites(ctx context.Context, uid int, values map[string]interface{}) error {
	if 0 == len(values) {
		return nil
	}
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil
	}
	if UserHasGroupXML(ctx, uid, "base.group_system") {
		return nil
	}
	denyFields := writeDenyUnlessSysFields("core.user")
	for k := range values {
		if k == "id" {
			continue
		}
		if denyFields[k] {
			return fmt.Errorf("field %q on core.user requires system administrator", k)
		}
	}
	return nil
}
