package render

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

type BreadcrumbItem struct {
	Label string
	Href  string // empty = current page (no link)
}

// HomeWebURL returns the canonical Home dashboard URL.
func HomeWebURL(ctx context.Context) string {
	return "/web/home"
}

// SettingsHomeURL is the canonical Settings hub URL.
func SettingsHomeURL() string {
	return "/web/settings"
}

// MenuWebURL builds a /web URL for a sys.menu row.
func MenuWebURL(menuID, actionID int) string {
	return WorkspaceURL(WorkspaceQuery{ActionID: actionID, MenuID: strconv.Itoa(menuID)})
}

// listViewURL returns the list view URL for a form workspace query string.
func listViewURL(formBaseQuery string) string {
	formBaseQuery = strings.TrimSpace(formBaseQuery)
	if formBaseQuery == "" {
		return ""
	}
	u, err := url.ParseQuery(formBaseQuery)
	if err != nil {
		return ""
	}
	u.Set(WorkspaceViewTypeParam, ViewModeList)
	u.Del(WorkspaceRecordIDParam)
	return WorkspaceRoute + "?" + u.Encode()
}

type menuCrumb struct {
	ID       int
	Name     string
	ActionID int
}

func collectMenuAncestors(ctx context.Context, leafID int) []menuCrumb {
	var stack []menuCrumb
	id := leafID
	seen := make(map[int]bool)
	for id > 0 && !seen[id] {
		seen[id] = true
		row, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"id": id})
		if err != nil {
			break
		}
		name := strings.TrimSpace(orm.AsString(row["name"]))
		aid, _ := orm.CoerceInt64(row["action_id"])
		stack = append(stack, menuCrumb{ID: id, Name: name, ActionID: int(aid)})
		pid := 0
		if pv, ok := row["parent_id"]; ok {
			if v, ok2 := orm.CoerceInt64(pv); ok2 && v > 0 {
				pid = int(v)
			}
		}
		id = pid
	}
	// leaf was pushed first; reverse to root→leaf
	for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
		stack[i], stack[j] = stack[j], stack[i]
	}
	return stack
}

// splitModuleMenuChain separates the application root menu from descendant menu folders/items.
func splitModuleMenuChain(chain []menuCrumb) (module menuCrumb, menus []menuCrumb, ok bool) {
	if len(chain) == 0 {
		return menuCrumb{}, nil, false
	}
	return chain[0], chain[1:], true
}

type BreadcrumbInput struct {
	ActiveMenuID  string
	ResModel      string
	ViewType      string
	FormBaseQuery string
	Record        map[string]interface{}
	RecordID      int
}

// BuildWorkspaceBreadcrumbs builds module → menu → view → record for workspace pages.
// Labels come from sys.menu names, generic view-mode names, and record/model metadata — never addon-specific maps in core.
func BuildWorkspaceBreadcrumbs(ctx context.Context, in BreadcrumbInput) []BreadcrumbItem {
	mid, err := strconv.Atoi(strings.TrimSpace(in.ActiveMenuID))
	if err != nil || mid <= 0 {
		if label := workspaceRecordBreadcrumbLabel(in); label != "" {
			return []BreadcrumbItem{{Label: label, Href: ""}}
		}
		if viewLabel := workspaceViewBreadcrumbLabel(in.ViewType); viewLabel != "" {
			return []BreadcrumbItem{{Label: viewLabel, Href: ""}}
		}
		return nil
	}

	chain := collectMenuAncestors(ctx, mid)
	module, menus, ok := splitModuleMenuChain(chain)
	if !ok {
		return nil
	}

	vt := strings.ToLower(strings.TrimSpace(in.ViewType))
	isMatrix := vt == ViewModeList || vt == ViewModeKanban
	settingsRootID := settingsRootMenuID(ctx, in.ActiveMenuID)

	var items []BreadcrumbItem

	// Module (application root menu, e.g. CRM).
	moduleHref := menuCrumbHref(module, settingsRootID, false)
	items = append(items, BreadcrumbItem{Label: module.Name, Href: moduleHref})

	// Menu path (folders and leaf action menu below the module root).
	allMenus := append([]menuCrumb{}, menus...)
	for i, m := range allMenus {
		isLastMenu := i == len(allMenus)-1
		href := menuCrumbHref(m, settingsRootID, false)
		if isLastMenu && isMatrix {
			href = ""
		}
		if isLastMenu && vt == ViewModeForm && in.RecordID > 0 {
			if listHref := listViewURL(in.FormBaseQuery); listHref != "" {
				href = listHref
			}
		}
		items = append(items, BreadcrumbItem{Label: m.Name, Href: href})
	}

	// View (non-primary modes such as graph/pivot; list/kanban/form omit when menu names the screen).
	if viewLabel := workspaceViewBreadcrumbLabel(in.ViewType); viewLabel != "" {
		if len(items) == 0 || items[len(items)-1].Label != viewLabel {
			items = append(items, BreadcrumbItem{Label: viewLabel, Href: ""})
		}
	}

	// Record (existing row name, or new-form model label from generic UIModelName fallback).
	if recordLabel := workspaceRecordBreadcrumbLabel(in); recordLabel != "" {
		if len(items) == 0 || items[len(items)-1].Label != recordLabel {
			items = append(items, BreadcrumbItem{Label: recordLabel, Href: ""})
		}
	}

	return items
}

func settingsRootMenuID(ctx context.Context, activeMenuID string) int {
	if !IsMenuUnderSettingsRoot(ctx, activeMenuID) {
		return 0
	}
	rid, _, err := orm.ResolveXmlId(ctx, "base.menu_settings_root")
	if err != nil || rid <= 0 {
		return 0
	}
	return rid
}

func menuCrumbHref(m menuCrumb, settingsRootID int, current bool) string {
	if current {
		return ""
	}
	href := MenuWebURL(m.ID, m.ActionID)
	if settingsRootID > 0 && m.ID == settingsRootID {
		return SettingsHomeURL()
	}
	return href
}

// workspaceViewBreadcrumbLabel returns a view-mode segment for auxiliary workspace modes.
func workspaceViewBreadcrumbLabel(viewType string) string {
	vt := strings.ToLower(strings.TrimSpace(viewType))
	switch vt {
	case ViewModeList, ViewModeKanban, ViewModeForm, "":
		return ""
	default:
		if suffix, ok := viewModeSuffix[vt]; ok {
			return suffix
		}
		return ""
	}
}

// workspaceRecordBreadcrumbLabel returns the record segment for form views.
func workspaceRecordBreadcrumbLabel(in BreadcrumbInput) string {
	vt := strings.ToLower(strings.TrimSpace(in.ViewType))
	if vt != ViewModeForm {
		return ""
	}
	if in.RecordID > 0 {
		label := ""
		if in.Record != nil {
			label = strings.TrimSpace(recStr(in.Record, "name"))
		}
		if label == "" {
			label = "Record"
		}
		return label
	}
	return UIModelName(strings.TrimSpace(in.ResModel))
}

// BuildAppsBreadcrumbs returns Home + Apps (+ optional module detail as current).
func BuildAppsBreadcrumbs(ctx context.Context, appsListHref string, detailTitle string) []BreadcrumbItem {
	out := []BreadcrumbItem{
		{Label: "Home", Href: HomeWebURL(ctx)},
	}
	if strings.TrimSpace(detailTitle) != "" {
		listHref := strings.TrimSpace(appsListHref)
		if listHref == "" {
			listHref = "/web/apps"
		}
		out = append(out, BreadcrumbItem{Label: "Apps", Href: listHref})
		out = append(out, BreadcrumbItem{Label: strings.TrimSpace(detailTitle), Href: ""})
		return out
	}
	out = append(out, BreadcrumbItem{Label: "Apps", Href: ""})
	return out
}

// BuildHomeDashboardBreadcrumbs returns Home + Dashboard (current).
func BuildHomeDashboardBreadcrumbs(ctx context.Context) []BreadcrumbItem {
	return []BreadcrumbItem{
		{Label: "Home", Href: HomeWebURL(ctx)},
		{Label: "Dashboard", Href: ""},
	}
}

// BuildSettingsHubBreadcrumbs returns a single Settings crumb for the hub page.
func BuildSettingsHubBreadcrumbs(ctx context.Context) []BreadcrumbItem {
	return []BreadcrumbItem{
		{Label: "Settings", Href: ""},
	}
}

// BuildAppLogsBreadcrumbs returns Home + menu path to the Event Log menu (current = leaf).
func BuildAppLogsBreadcrumbs(ctx context.Context, appLogsMenuID int) []BreadcrumbItem {
	items := []BreadcrumbItem{
		{Label: "Home", Href: HomeWebURL(ctx)},
	}
	if appLogsMenuID <= 0 {
		return items
	}
	chain := collectMenuAncestors(ctx, appLogsMenuID)
	for i, m := range chain {
		isLast := i == len(chain)-1
		href := MenuWebURL(m.ID, m.ActionID)
		if isLast {
			href = ""
		}
		items = append(items, BreadcrumbItem{Label: m.Name, Href: href})
	}
	return items
}
