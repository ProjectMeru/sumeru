package web

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// ResolveWindowActionID returns sys.action.window id from query action (numeric or xml id) and/or menu_id.
func ResolveWindowActionID(ctx context.Context, actionQuery, menuQuery string) int {
	if actionID := resolveActionIDFromQuery(ctx, actionQuery); actionID != 0 {
		return actionID
	}
	return resolveActionIDFromMenuQuery(ctx, menuQuery)
}

func resolveActionIDFromQuery(ctx context.Context, actionQuery string) int {
	actionQuery = strings.TrimSpace(actionQuery)
	if actionQuery == "" {
		return 0
	}
	if actionID, err := strconv.Atoi(actionQuery); err == nil {
		return actionID
	}
	actionID, _, err := orm.ResolveXmlId(ctx, actionQuery)
	if err != nil {
		return 0
	}
	return actionID
}

func resolveActionIDFromMenuQuery(ctx context.Context, menuQuery string) int {
	menuID, ok := parseMenuIDString(menuQuery)
	if !ok {
		return 0
	}
	return windowActionIDFromMenu(ctx, menuID)
}

func windowActionIDFromMenu(ctx context.Context, menuID int) int {
	menuRecord, err := orm.SearchOne(ctx, sysMenuModel, map[string]interface{}{"id": menuID})
	if err != nil {
		return 0
	}
	if actionID, ok := menuRecordActionID(menuRecord); ok {
		return actionID
	}
	return firstDescendantWindowActionID(ctx, menuID)
}

func menuRecordActionID(menuRecord map[string]interface{}) (actionID int, ok bool) {
	actionID64, hasAction := orm.CoerceInt64(menuRecord["action_id"])
	if !hasAction || actionID64 == 0 {
		return 0, false
	}
	return int(actionID64), true
}

// firstDescendantWindowActionID returns the first non-zero action_id in a depth-first walk
// of children ordered by sequence, then id.
func firstDescendantWindowActionID(ctx context.Context, parentMenuID int) int {
	menuTable := orm.MustQuotedTableName(sysMenuModel)
	rows, err := orm.DB.QueryContext(ctx,
		`SELECT id, action_id FROM `+menuTable+` WHERE parent_id = $1 ORDER BY sequence ASC, id ASC`,
		parentMenuID,
	)
	if err != nil {
		return 0
	}
	defer rows.Close()

	for rows.Next() {
		var childMenuID int
		var childActionID sql.NullInt64
		if err := rows.Scan(&childMenuID, &childActionID); err != nil {
			continue
		}
		if childActionID.Valid && childActionID.Int64 != 0 {
			return int(childActionID.Int64)
		}
		if descendantActionID := firstDescendantWindowActionID(ctx, childMenuID); descendantActionID != 0 {
			return descendantActionID
		}
	}
	return 0
}

func menuIDPointsToAppLogs(ctx context.Context, menuQuery string) bool {
	return menuIDMatchesXML(ctx, menuQuery, appLogsMenuXMLID)
}

func menuIDMatchesXML(ctx context.Context, menuQuery, menuXMLID string) bool {
	expectedMenuID, ok := resolvedMenuIDFromXML(ctx, menuXMLID)
	if !ok {
		return false
	}
	actualMenuID, ok := parseMenuIDString(menuQuery)
	if !ok {
		return false
	}
	return actualMenuID == expectedMenuID
}

// isHomeMenuTree reports whether menu_id is base.menu_home_root or a descendant.
// An empty menu_id is treated as the home root (default landing behavior).
func isHomeMenuTree(ctx context.Context, menuQuery string) bool {
	if strings.TrimSpace(menuQuery) == "" {
		return true
	}
	return menuIsUnderXMLRoot(ctx, menuQuery, homeMenuRootXMLID)
}

func menuIsUnderXMLRoot(ctx context.Context, menuQuery, rootMenuXMLID string) bool {
	rootMenuID, ok := resolvedMenuIDFromXML(ctx, rootMenuXMLID)
	if !ok {
		return false
	}
	menuID, ok := parseMenuIDString(menuQuery)
	if !ok {
		return false
	}
	return orm.MenuHasAncestor(ctx, menuID, rootMenuID)
}

func resolvedMenuIDFromXML(ctx context.Context, menuXMLID string) (menuID int, ok bool) {
	menuID, _, err := orm.ResolveXmlId(ctx, menuXMLID)
	if err != nil || menuID <= 0 {
		return 0, false
	}
	return menuID, true
}

func parseMenuIDString(menuQuery string) (menuID int, ok bool) {
	menuID, err := strconv.Atoi(strings.TrimSpace(menuQuery))
	if err != nil || menuID <= 0 {
		return 0, false
	}
	return menuID, true
}

// actionWindowTargetModel returns the ORM technical model for a sys.action.window row (core_model).
func actionWindowTargetModel(actionData map[string]interface{}) string {
	return strings.TrimSpace(orm.AsString(actionData["core_model"]))
}

// CanonicalMenuID rewrites folder/root menu_id to the leaf menu that owns actionID.
func CanonicalMenuID(ctx context.Context, menuQuery string, actionID int) string {
	menuID, ok := parseMenuIDString(menuQuery)
	if !ok || actionID == 0 {
		return menuQuery
	}
	menuRecord, err := orm.SearchOne(ctx, sysMenuModel, map[string]interface{}{"id": menuID})
	if err != nil {
		return menuQuery
	}
	if aid, ok := menuRecordActionID(menuRecord); ok && aid == actionID {
		return menuQuery
	}
	if leafID := menuIDForWindowAction(ctx, menuID, actionID); leafID > 0 {
		return strconv.Itoa(leafID)
	}
	return menuQuery
}

func menuIDForWindowAction(ctx context.Context, parentMenuID, actionID int) int {
	if parentMenuID <= 0 || actionID == 0 {
		return 0
	}
	menuTable := orm.MustQuotedTableName(sysMenuModel)
	rows, err := orm.DB.QueryContext(ctx,
		`SELECT id, action_id FROM `+menuTable+` WHERE parent_id = $1 ORDER BY sequence ASC, id ASC`,
		parentMenuID,
	)
	if err != nil {
		return 0
	}
	defer rows.Close()

	for rows.Next() {
		var childMenuID int
		var childActionID sql.NullInt64
		if err := rows.Scan(&childMenuID, &childActionID); err != nil {
			continue
		}
		if childActionID.Valid && int(childActionID.Int64) == actionID {
			return childMenuID
		}
		if found := menuIDForWindowAction(ctx, childMenuID, actionID); found > 0 {
			return found
		}
	}
	return 0
}
