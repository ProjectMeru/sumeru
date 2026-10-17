package orm

import (
	"context"
	"strconv"
	"strings"
)

const sysMenuModel = "sys.menu"

type menuNavRow struct {
	id           int
	actionID     int
	accessGroups string
}

// MenuRecordActionID returns a non-zero sys.menu action_id when present.
func MenuRecordActionID(menuRecord map[string]interface{}) (actionID int, ok bool) {
	actionID64, hasAction := CoerceInt64(menuRecord["action_id"])
	if !hasAction || 0 == actionID64 {
		return 0, false
	}
	return int(actionID64), true
}

// WindowActionIDFromMenu resolves a window action id from a menu row or its descendants.
func WindowActionIDFromMenu(ctx context.Context, menuID int) int {
	actionID, _ := MenuEntryAction(ctx, menuID)
	return actionID
}

// MenuEntryAction returns the action_id and the menu row id that owns it (leaf when resolved from a folder).
func MenuEntryAction(ctx context.Context, menuID int) (actionID int, leafMenuID int) {
	if 0 >= menuID || DB == nil {
		return 0, 0
	}
	row, ok := loadMenuNavRow(ctx, menuID)
	if !ok {
		return 0, 0
	}
	if row.actionID != 0 {
		return row.actionID, menuID
	}
	for _, child := range loadMenuNavChildren(ctx, menuID) {
		if aid, leaf := MenuEntryAction(ctx, child.id); aid != 0 {
			return aid, leaf
		}
	}
	return 0, 0
}

// MenuEntryActionForUser returns the first window action under menuID the user may open.
func MenuEntryActionForUser(ctx context.Context, uid, menuID int) (actionID int, leafMenuID int) {
	if 0 >= menuID || DB == nil {
		return 0, 0
	}
	return menuEntryActionForUser(ctx, uid, menuID)
}

func menuEntryActionForUser(ctx context.Context, uid, menuID int) (actionID int, leafMenuID int) {
	row, ok := loadMenuNavRow(ctx, menuID)
	if !ok {
		return 0, 0
	}
	if row.actionID != 0 && UserMayAccessMenu(ctx, uid, row.accessGroups) {
		return row.actionID, menuID
	}
	for _, child := range loadMenuNavChildren(ctx, menuID) {
		if aid, leaf := menuEntryActionForUser(ctx, uid, child.id); aid != 0 {
			return aid, leaf
		}
	}
	return 0, 0
}

// MenuIDForWindowAction finds the descendant menu row that binds actionID under parentMenuID.
func MenuIDForWindowAction(ctx context.Context, parentMenuID, actionID int) int {
	if 0 >= parentMenuID || 0 == actionID {
		return 0
	}
	_, leafID := walkMenuChildren(ctx, parentMenuID, func(childMenuID int, _ map[string]interface{}, childActionID int) (int, int, bool) {
		if childActionID == actionID {
			return actionID, childMenuID, true
		}
		if leaf := MenuIDForWindowAction(ctx, childMenuID, actionID); leaf > 0 {
			return actionID, leaf, true
		}
		return 0, 0, false
	})
	return leafID
}

// CanonicalMenuID rewrites folder/root menu_id to the leaf menu that owns actionID.
func CanonicalMenuID(ctx context.Context, menuQuery string, actionID int) string {
	menuID, ok := ParseMenuIDString(menuQuery)
	if !ok || 0 == actionID {
		return menuQuery
	}
	row, err := SearchOne(ctx, sysMenuModel, map[string]interface{}{"id": menuID})
	if err != nil {
		return menuQuery
	}
	if aid, ok := MenuRecordActionID(row); ok && aid == actionID {
		return menuQuery
	}
	if leafID := MenuIDForWindowAction(ctx, menuID, actionID); leafID > 0 {
		return strconv.Itoa(leafID)
	}
	return menuQuery
}

// ParseMenuIDString parses a positive numeric menu_id query value.
func ParseMenuIDString(menuQuery string) (menuID int, ok bool) {
	menuID, err := strconv.Atoi(strings.TrimSpace(menuQuery))
	if err != nil || 0 >= menuID {
		return 0, false
	}
	return menuID, true
}

func loadMenuNavRow(ctx context.Context, menuID int) (menuNavRow, bool) {
	table := MustQuotedTableName(sysMenuModel)
	q := `SELECT id, COALESCE(action_id, 0), COALESCE(NULLIF(TRIM(access_groups), ''), '')
		FROM ` + table + ` WHERE id = $1`
	var row menuNavRow
	err := DB.QueryRowContext(ctx, q, menuID).Scan(&row.id, &row.actionID, &row.accessGroups)
	if err != nil {
		return menuNavRow{}, false
	}
	return row, true
}

func loadMenuNavChildren(ctx context.Context, parentMenuID int) []menuNavRow {
	table := MustQuotedTableName(sysMenuModel)
	q := `SELECT id, COALESCE(action_id, 0), COALESCE(NULLIF(TRIM(access_groups), ''), '')
		FROM ` + table + ` WHERE parent_id = $1 ORDER BY sequence ASC, id ASC`
	rows, err := DB.QueryContext(ctx, q, parentMenuID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]menuNavRow, 0, 8)
	for rows.Next() {
		var row menuNavRow
		if err := rows.Scan(&row.id, &row.actionID, &row.accessGroups); err != nil {
			continue
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return out
}

func walkMenuChildren(
	ctx context.Context,
	parentMenuID int,
	visit func(childMenuID int, childRow map[string]interface{}, childActionID int) (actionID int, menuID int, ok bool),
) (actionID int, menuID int) {
	for _, child := range loadMenuNavChildren(ctx, parentMenuID) {
		childMenuID := child.id
		row := map[string]interface{}{
			"id":            childMenuID,
			"action_id":     child.actionID,
			"access_groups": child.accessGroups,
		}
		if a, m, ok := visit(childMenuID, row, child.actionID); ok {
			return a, m
		}
	}
	return 0, 0
}
