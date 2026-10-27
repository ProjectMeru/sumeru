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

// UserMayAccessMenu applies fail-closed menu policy for workspace gates: empty access_groups
// is visible only to base.group_system (and superuser/bypass). Tagged menus require a matching group.
func UserMayAccessMenu(ctx context.Context, uid int, accessGroupsCSV string) bool {
	if menuAccessBypass(ctx, uid) {
		return true
	}
	ids := NormalizeAccessGroupList(accessGroupsCSV)
	if 0 == len(ids) {
		return UserHasGroupXML(ctx, uid, "base.group_system")
	}
	return userMatchesAnyGroupXML(ctx, uid, ids)
}

// UserMayAccessMenuNavigation is used for shell, launcher, and menu→action resolution.
// Settings administrators (base.group_system) may reach any menu row for navigation.
func UserMayAccessMenuNavigation(ctx context.Context, uid int, accessGroupsCSV string) bool {
	if menuAccessBypass(ctx, uid) || UserHasGroupXML(ctx, uid, "base.group_system") {
		return true
	}
	ids := NormalizeAccessGroupList(accessGroupsCSV)
	if 0 == len(ids) {
		return false
	}
	return userMatchesAnyGroupXML(ctx, uid, ids)
}

func menuAccessBypass(ctx context.Context, uid int) bool {
	return SecurityBypass(ctx) || uid == superuserUID
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
	return keepOneExclusiveGroup(ctx, out)
}

func keepOneExclusiveGroup(ctx context.Context, groupIDs []int) []int {
	if DB == nil || 0 == len(groupIDs) {
		return groupIDs
	}
	placeholders := make([]string, len(groupIDs))
	args := make([]interface{}, len(groupIDs))
	for i, id := range groupIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	q := `SELECT g.id, COALESCE(g.category_id, 0), COALESCE(c.exclusive, false)
		FROM ` + MustQuotedTableName("core.group") + ` g
		LEFT JOIN ` + MustQuotedTableName("sys.module.category") + ` c ON c.id = g.category_id
		WHERE g.id IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := DB.QueryContext(ctx, q, args...)
	if err != nil {
		return groupIDs
	}
	defer rows.Close()
	exclusiveCategory := map[int]int{}
	for rows.Next() {
		var gid, categoryID int
		var exclusive bool
		if err := rows.Scan(&gid, &categoryID, &exclusive); err != nil {
			return groupIDs
		}
		if exclusive && categoryID > 0 {
			exclusiveCategory[gid] = categoryID
		}
	}
	if err := rows.Err(); err != nil {
		return groupIDs
	}
	return dropExtraExclusiveGroups(groupIDs, exclusiveCategory)
}

func dropExtraExclusiveGroups(groupIDs []int, exclusiveCategory map[int]int) []int {
	if 0 == len(exclusiveCategory) {
		return groupIDs
	}
	seenCategory := map[int]struct{}{}
	out := make([]int, 0, len(groupIDs))
	for _, gid := range groupIDs {
		categoryID, exclusive := exclusiveCategory[gid]
		if !exclusive {
			out = append(out, gid)
			continue
		}
		if _, taken := seenCategory[categoryID]; taken {
			continue
		}
		seenCategory[categoryID] = struct{}{}
		out = append(out, gid)
	}
	return out
}

// MergeAccessGroupCSV applies menu group tokens. A leading '-' removes that XML id
// from the stored list. Without a removal token the incoming list replaces the stored one.
func MergeAccessGroupCSV(existing, incoming string) string {
	tokens := NormalizeAccessGroupList(incoming)
	var remove, add []string
	for _, token := range tokens {
		if strings.HasPrefix(token, "-") {
			if id := strings.TrimPrefix(token, "-"); id != "" {
				remove = append(remove, id)
			}
			continue
		}
		add = append(add, token)
	}
	if 0 == len(remove) {
		return strings.Join(add, ",")
	}
	drop := make(map[string]struct{}, len(remove))
	for _, id := range remove {
		drop[id] = struct{}{}
	}
	seen := make(map[string]struct{})
	var out []string
	for _, id := range NormalizeAccessGroupList(existing) {
		if _, skip := drop[id]; skip {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, id := range add {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return strings.Join(out, ",")
}

// ListAllGroupRows returns id,name for UI pickers.
func ListAllGroupRows(ctx context.Context) ([]map[string]interface{}, error) {
	return Search(ctx, "core.group", nil)
}
