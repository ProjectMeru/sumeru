package orm

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/applog"
)

// ApplyUserSecurityPost applies core.user security side effects from a form POST
// (companies, groups, password) after the main record save.
func ApplyUserSecurityPost(ctx context.Context, actor, userID int, form url.Values) {
	if userID <= 0 {
		return
	}
	if err := CheckModelAccess(ctx, actor, "core.user", "write"); err != nil {
		return
	}
	if _, ok := form["company_ids"]; ok {
		if !UserHasGroupXML(ctx, actor, "base.group_system") {
			applog.WarnMsg(ctx, "orm", "user_security", "deny set companies: actor not system admin", nil,
				map[string]interface{}{"user_id": userID, "actor": actor})
		} else {
			var cids []int
			for _, s := range form["company_ids"] {
				n, err := strconv.Atoi(strings.TrimSpace(s))
				if err == nil && n > 0 {
					cids = append(cids, n)
				}
			}
			if err := SetUserCompanyLinks(ctx, userID, cids); err != nil {
				applog.WarnMsg(ctx, "orm", "user_security", "set user companies failed", err,
					map[string]interface{}{"user_id": userID})
			}
		}
	}
	if form.Get("security_groups_touched") == "1" {
		if !UserHasGroupXML(ctx, actor, "base.group_system") {
			applog.WarnMsg(ctx, "orm", "user_security", "deny set groups: actor not system admin", nil,
				map[string]interface{}{"user_id": userID, "actor": actor})
			return
		}
		var gids []int
		if ut := strings.TrimSpace(form.Get("security_user_type")); ut != "" {
			if n, err := strconv.Atoi(ut); err == nil && n > 0 {
				gids = append(gids, n)
			}
		}
		for _, s := range form["security_group_ids"] {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			if err == nil && n > 0 {
				gids = append(gids, n)
			}
		}
		if err := SetUserGroupLinks(ctx, userID, gids); err != nil {
			applog.WarnMsg(ctx, "orm", "user_security", "set user groups failed", err,
				map[string]interface{}{"user_id": userID})
		}
	}
	if _, ok := form["password_plain"]; ok {
		if pw := strings.TrimSpace(form.Get("password_plain")); pw != "" {
			confirm := strings.TrimSpace(form.Get("password_plain_confirm"))
			if pw != confirm {
				applog.WarnMsg(ctx, "orm", "user_security", "password confirm mismatch", nil,
					map[string]interface{}{"user_id": userID})
				return
			}
			if err := SetUserPassword(ctx, actor, userID, pw); err != nil {
				applog.WarnMsg(ctx, "orm", "user_security", "password update failed", err,
					map[string]interface{}{"user_id": userID})
			}
		}
	}
}
