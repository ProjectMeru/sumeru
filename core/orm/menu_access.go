package orm

import (
	"context"
	"fmt"
	"strings"
)

// UserHasAnyAccessGroup resolves comma-separated XML ids for view/field gates.
// Empty accessGroups means unrestricted (caller already reached the form/view).
func UserHasAnyAccessGroup(ctx context.Context, uid int, accessGroupsCSV string) bool {
	if SecurityBypass(ctx) || uid == superuserUID {
		return true
	}
	ids := NormalizeAccessGroupList(accessGroupsCSV)
	if 0 == len(ids) {
		return true
	}
	return userMatchesAnyGroupXML(ctx, uid, ids)
}

// UserMayAccessMenu applies fail-closed menu policy: empty access_groups is visible
// only to base.group_system (and superuser/bypass). Tagged menus require a matching group.
func UserMayAccessMenu(ctx context.Context, uid int, accessGroupsCSV string) bool {
	if SecurityBypass(ctx) || uid == superuserUID {
		return true
	}
	ids := NormalizeAccessGroupList(accessGroupsCSV)
	if 0 == len(ids) {
		return UserHasGroupXML(ctx, uid, "base.group_system")
	}
	return userMatchesAnyGroupXML(ctx, uid, ids)
}

// UserHasGroupXML reports whether uid's effective groups include the given XML id.
func UserHasGroupXML(ctx context.Context, uid int, xmlID string) bool {
	if SecurityBypass(ctx) || uid == superuserUID {
		return true
	}
	xmlID = strings.TrimSpace(xmlID)
	if xmlID == "" || 0 >= uid {
		return false
	}
	return userMatchesAnyGroupXML(ctx, uid, []string{xmlID})
}

func userMatchesAnyGroupXML(ctx context.Context, uid int, xmlIDs []string) bool {
	if 0 >= uid {
		return false
	}
	eff, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return false
	}
	if 0 == len(eff) {
		return false
	}

	for _, xmlID := range xmlIDs {
		gid, _, err := ResolveXmlId(ctx, xmlID)
		if err != nil || 0 == gid {
			continue
		}
		if intSliceContains(eff, gid) {
			return true
		}
	}
	return false
}

// SetUserGroupLinks replaces group membership for a user (explicit rel rows only).
// Enforces mutual exclusion among base.group_user / group_portal / group_public.
func SetUserGroupLinks(ctx context.Context, userID int, groupIDs []int) error {
	if DB == nil {
		return fmt.Errorf("no database")
	}
	groupIDs = normalizeUserTypeGroupIDs(ctx, groupIDs)
	tbl := MustQuotedTableName(tableGroupUserRel)
	if _, err := DB.ExecContext(ctx, `DELETE FROM `+tbl+` WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, gid := range groupIDs {
		if 0 >= gid {
			continue
		}
		if _, err := DB.ExecContext(ctx, `INSERT INTO `+tbl+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT (user_id, group_id) DO NOTHING`, userID, gid); err != nil {
			return err
		}
	}
	InvalidateEffectiveGroups(ctx, userID)
	return nil
}

func normalizeUserTypeGroupIDs(ctx context.Context, groupIDs []int) []int {
	typeXML := []string{"base.group_user", "base.group_portal", "base.group_public"}
	typeIDs := map[int]string{}
	for _, x := range typeXML {
		gid, _, err := ResolveXmlId(ctx, x)
		if err == nil && gid > 0 {
			typeIDs[gid] = x
		}
	}
	chosen := ""
	var out []int
	seen := map[int]struct{}{}
	for _, gid := range groupIDs {
		if 0 >= gid {
			continue
		}
		if x, ok := typeIDs[gid]; ok {
			if chosen == "" {
				chosen = x
				out = append(out, gid)
				seen[gid] = struct{}{}
			}
			continue
		}
		if _, ok := seen[gid]; ok {
			continue
		}
		seen[gid] = struct{}{}
		out = append(out, gid)
	}
	return out
}

// ListAllGroupRows returns id,name for UI pickers.
func ListAllGroupRows(ctx context.Context) ([]map[string]interface{}, error) {
	return Search(ctx, "core.group", nil)
}
