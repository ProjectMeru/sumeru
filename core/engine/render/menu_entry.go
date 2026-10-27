package render

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// ModuleWorkspaceEntryURL is the user-facing link to open an application module (/web/{name}).
func ModuleWorkspaceEntryURL(ctx context.Context, moduleName string) string {
	_ = ctx
	if href := ModuleWebURL(moduleName); href != "" {
		return href
	}
	return HomeWebURL()
}

// MenuWorkspaceEntryURL builds the internal workspace URL for a menu row (action + menu_id).
func MenuWorkspaceEntryURL(ctx context.Context, menuID int) string {
	if 0 >= menuID {
		return HomeWebURL()
	}
	var actionID, leafMenuID int
	if uid := orm.SecurityUID(ctx); uid > 0 {
		actionID, leafMenuID = orm.MenuEntryActionForUser(ctx, uid, menuID)
	} else {
		actionID, leafMenuID = orm.MenuEntryAction(ctx, menuID)
	}
	return WorkspaceURLFromMenuAction(ctx, actionID, menuID, leafMenuID)
}

// WorkspaceURLFromMenuAction builds the internal workspace URL from a resolved menu entry action.
func WorkspaceURLFromMenuAction(ctx context.Context, actionID, rootMenuID, leafMenuID int) string {
	if actionID == 0 {
		return HomeWebURL()
	}
	menuIDStr := strconv.Itoa(rootMenuID)
	if leafMenuID > 0 {
		menuIDStr = strconv.Itoa(leafMenuID)
	}
	if href := urlActionHref(ctx, actionID); href != "" {
		return href
	}
	return WorkspaceURL(WorkspaceQuery{ActionID: actionID, MenuID: menuIDStr})
}

// WorkspaceURLWithDebug appends debug=true to a workspace URL when enabled.
func WorkspaceURLWithDebug(workspaceURL string, debugEnabled bool) string {
	workspaceURL = strings.TrimSpace(workspaceURL)
	if !debugEnabled || workspaceURL == "" {
		return workspaceURL
	}
	if strings.Contains(workspaceURL, "debug=") {
		return workspaceURL
	}
	sep := "?"
	if strings.Contains(workspaceURL, "?") {
		sep = "&"
	}
	return workspaceURL + sep + "debug=true"
}

func urlActionHref(ctx context.Context, actionID int) string {
	row, err := orm.SearchOne(ctx, "sys.action.url", map[string]interface{}{"id": actionID})
	if err != nil {
		return ""
	}
	url := strings.TrimSpace(orm.AsString(row["url"]))
	if url == "" || !strings.HasPrefix(url, "/") {
		return ""
	}
	return url
}

func menuItemWorkspaceHref(ctx context.Context, menuRowID int, directActionID int, urlHrefs map[int]string, moduleName string, isRoot bool) string {
	if isRoot && strings.TrimSpace(moduleName) != "" {
		if href := ModuleWebURL(moduleName); href != "" {
			return href
		}
	}
	if directActionID != 0 {
		if href := urlHrefs[directActionID]; href != "" {
			return href
		}
		return fmt.Sprintf("/web?action=%d&menu_id=%d", directActionID, menuRowID)
	}
	return MenuWorkspaceEntryURL(ctx, menuRowID)
}
