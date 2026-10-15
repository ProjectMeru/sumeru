package orm

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/applog"
)

// ApplyUserSecurityPost applies core.user security side effects from a form POST
// (companies, groups, password) after the main record save.
func ApplyUserSecurityPost(ctx context.Context, actor, userID int, form url.Values) {
	_ = ApplyUserSecurityPostErr(ctx, actor, userID, form)
}

// ApplyUserSecurityPostErr is like ApplyUserSecurityPost but returns the first hard error.
func ApplyUserSecurityPostErr(ctx context.Context, actor, userID int, form url.Values) error {
	if 0 >= userID {
		return fmt.Errorf("invalid user id")
	}
	if err := CheckModelAccess(ctx, actor, "core.user", "write"); err != nil {
		return err
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
				return fmt.Errorf("set user companies: %w", err)
			}
		}
	}
	if form.Get("security_groups_touched") == "1" {
		if !UserHasGroupXML(ctx, actor, "base.group_system") {
			return fmt.Errorf("access denied")
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
			return fmt.Errorf("set user groups: %w", err)
		}
	}
	if _, ok := form["password_plain"]; ok {
		if pw := strings.TrimSpace(form.Get("password_plain")); pw != "" {
			confirm := strings.TrimSpace(form.Get("password_plain_confirm"))
			if pw != confirm {
				return fmt.Errorf("password confirmation does not match")
			}
			if err := SetUserPassword(ctx, actor, userID, pw); err != nil {
				return err
			}
		}
	}
	return nil
}
