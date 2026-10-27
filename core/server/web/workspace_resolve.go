package web

import (
	"context"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// ResolveWindowActionID returns sys.action.window id from query action (numeric or xml id) and/or menu_id.
func ResolveWindowActionID(ctx context.Context, actionQuery, menuQuery string) int {
	if actionID := resolveActionIDFromQuery(ctx, actionQuery); actionID != 0 {
		return actionID
	}
	menuID, ok := orm.ParseMenuIDString(menuQuery)
	if !ok {
		return 0
	}
	actionID, _ := orm.MenuEntryActionForUser(ctx, orm.SecurityUID(ctx), menuID)
	if actionID != 0 {
		return actionID
	}
	return orm.WindowActionIDFromMenu(ctx, menuID)
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

func menuRecordActionID(menuRecord map[string]interface{}) (actionID int, ok bool) {
	return orm.MenuRecordActionID(menuRecord)
}

func menuIDPointsToAppLogs(ctx context.Context, menuQuery string) bool {
	return menuIDMatchesXML(ctx, menuQuery, appLogsMenuXMLID)
}

func menuIDMatchesXML(ctx context.Context, menuQuery, menuXMLID string) bool {
	expectedMenuID, ok := resolvedMenuIDFromXML(ctx, menuXMLID)
	if !ok {
		return false
	}
	actualMenuID, ok := orm.ParseMenuIDString(menuQuery)
	if !ok {
		return false
	}
	return actualMenuID == expectedMenuID
}

// isHomeMenuTree reports whether menu_id is base.menu_home_root or a descendant.
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
	menuID, ok := orm.ParseMenuIDString(menuQuery)
	if !ok {
		return false
	}
	return orm.MenuHasAncestor(ctx, menuID, rootMenuID)
}

func resolvedMenuIDFromXML(ctx context.Context, menuXMLID string) (menuID int, ok bool) {
	menuID, _, err := orm.ResolveXmlId(ctx, menuXMLID)
	if err != nil || 0 >= menuID {
		return 0, false
	}
	return menuID, true
}

func parseMenuIDString(menuQuery string) (menuID int, ok bool) {
	return orm.ParseMenuIDString(menuQuery)
}

func actionWindowTargetModel(actionData map[string]interface{}) string {
	return strings.TrimSpace(orm.AsString(actionData["core_model"]))
}

// CanonicalMenuID rewrites folder/root menu_id to the leaf menu that owns actionID.
func CanonicalMenuID(ctx context.Context, menuQuery string, actionID int) string {
	return orm.CanonicalMenuID(ctx, menuQuery, actionID)
}
