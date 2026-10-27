package render

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

var menuIconKey = regexp.MustCompile(`^[a-z0-9-]+$`)

func shellMenuAllowed(ctx context.Context, uid int, mi parser.MenuItem) bool {
	return orm.UserMayAccessMenuNavigation(ctx, uid, mi.AccessGroups)
}

func sanitizeMenuIcon(s string) string {
	s = strings.TrimSpace(s)
	if menuIconKey.MatchString(s) {
		return s
	}
	return ""
}

func LoadShellMenus(ctx context.Context, activeMenuID string) (topMenus []parser.MenuItem, sidebarMenus []SidebarMenu, activeModuleID, shellModuleTitle string) {
	shellModuleTitle = AppDisplayName
	allMenus := fetchShellMenus(ctx)
	if 0 == len(allMenus) {
		return nil, nil, "", AppDisplayName
	}

	appMods, err := queryInstalledApplicationNames(ctx)
	if err != nil {
		applog.WarnMsg(ctx, "render", "menus", "installed application modules", err, nil)
	}

	uid := orm.UIDFromContext(ctx)
	menuAllowed := func(mi parser.MenuItem) bool { return shellMenuAllowed(ctx, uid, mi) }

	topMenus = buildTopBarMenus(allMenus, appMods, menuAllowed)
	activeModuleID = resolveActiveModuleID(allMenus, activeMenuID)
	shellModuleTitle = shellTitleForModule(allMenus, activeModuleID, shellModuleTitle)
	sidebarMenus = buildSidebarMenus(allMenus, activeModuleID, menuAllowed)
	return topMenus, sidebarMenus, activeModuleID, shellModuleTitle
}

func fetchShellMenus(ctx context.Context) []parser.MenuItem {
	modTbl := orm.MustQuotedTableName("sys.module")
	menuTbl := orm.MustQuotedTableName("sys.menu")
	query := fmt.Sprintf(
		`SELECT m.id, m.name, m.parent_id, m.action_id, m.sequence,
		        COALESCE(NULLIF(TRIM(m.module), ''), '') AS module,
		        COALESCE(NULLIF(TRIM(m.web_icon), ''), '') AS web_icon,
		        COALESCE(NULLIF(TRIM(m.access_groups), ''), '') AS access_groups
		   FROM %s m
		  WHERE m.active = true
		    AND (COALESCE(NULLIF(TRIM(m.module), ''), '') = ''
		     OR EXISTS (
		      SELECT 1 FROM %s im WHERE im.name = m.module AND im.state = 'installed' AND im.active = true
		    ))
		  ORDER BY m.sequence`,
		menuTbl, modTbl,
	)
	rows, err := orm.DB.QueryContext(ctx, query)
	if err != nil {
		applog.WarnMsg(ctx, "render", "menus", "Error fetching menus", err, nil)
		return nil
	}
	defer rows.Close()

	lang := orm.UserLang(ctx)

	urlActionHrefs := loadURLActionHrefs(ctx)

	var allMenus []parser.MenuItem
	for rows.Next() {
		var id int
		var name, mod, webIcon, accessGroups string
		var parentID, actionID, seq sql.NullInt64
		if err := rows.Scan(&id, &name, &parentID, &actionID, &seq, &mod, &webIcon, &accessGroups); err != nil {
			applog.WarnMsg(ctx, "render", "menus", "Error scanning menu row", err, nil)
			continue
		}

		isRoot := !parentID.Valid || parentID.Int64 == int64(id)
		directAction := 0
		if actionID.Valid && actionID.Int64 != 0 {
			directAction = int(actionID.Int64)
		}
		m := parser.MenuItem{
			ID:           fmt.Sprintf("%d", id),
			Name:         orm.Translate(ctx, lang, name),
			Sequence:     int(seq.Int64),
			Module:       mod,
			WebIcon:      sanitizeMenuIcon(webIcon),
			AccessGroups: strings.TrimSpace(accessGroups),
		}
		if parentID.Valid && parentID.Int64 != int64(id) {
			m.ParentID = fmt.Sprintf("%d", parentID.Int64)
		}
		if isRoot && strings.EqualFold(strings.TrimSpace(name), "Home") {
			m.Action = HomeWebURL()
		} else {
			m.Action = menuItemWorkspaceHref(ctx, id, directAction, urlActionHrefs, mod, isRoot)
		}
		allMenus = append(allMenus, m)
	}
	if err := rows.Err(); err != nil {
		applog.WarnMsg(ctx, "render", "menus", "Menu rows error", err, nil)
	}
	applySettingsNavHrefOverrides(ctx, allMenus)
	return allMenus
}

func loadURLActionHrefs(ctx context.Context) map[int]string {
	out := make(map[int]string)
	rows, err := orm.Search(ctx, "sys.action.url", nil)
	if err != nil {
		applog.WarnMsg(ctx, "render", "menus", "Error loading URL actions", err, nil)
		return out
	}
	for _, row := range rows {
		id, ok := orm.CoerceInt64(row["id"])
		if !ok || 0 >= id {
			continue
		}
		url := strings.TrimSpace(orm.AsString(row["url"]))
		if url == "" || !strings.HasPrefix(url, "/") {
			continue
		}
		out[int(id)] = url
	}
	return out
}

func buildTopBarMenus(allMenus []parser.MenuItem, appMods map[string]struct{}, menuAllowed func(parser.MenuItem) bool) []parser.MenuItem {
	var topMenus []parser.MenuItem
	for _, m := range allMenus {
		if m.ParentID != "" {
			continue
		}
		if !menuAllowed(m) {
			continue
		}
		mod := strings.TrimSpace(m.Module)
		if len(appMods) > 0 {
			_, ok := appMods[mod]
			if mod == "" || !ok {
				continue
			}
		}
		if strings.EqualFold(strings.TrimSpace(m.Name), "Settings") {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(m.Name), "Home") {
			continue
		}
		topMenus = append(topMenus, m)
	}
	return sortTopBarRootMenus(topMenus)
}

func resolveActiveModuleID(allMenus []parser.MenuItem, activeMenuID string) string {
	if activeMenuID == "" {
		return ""
	}
	for _, m := range allMenus {
		if m.ID != activeMenuID {
			continue
		}
		curr := m
		for curr.ParentID != "" {
			found := false
			for _, parent := range allMenus {
				if parent.ID == curr.ParentID {
					curr = parent
					found = true
					break
				}
			}
			if !found || curr.ParentID == "" {
				break
			}
		}
		return curr.ID
	}
	return ""
}

func shellTitleForModule(allMenus []parser.MenuItem, activeModuleID, fallback string) string {
	if activeModuleID == "" {
		return fallback
	}
	for _, m := range allMenus {
		if m.ID == activeModuleID {
			if t := strings.TrimSpace(m.Name); t != "" {
				return t
			}
			break
		}
	}
	return fallback
}

func menuItemHasWindowOrURLAction(mi parser.MenuItem) bool {
	action := strings.TrimSpace(mi.Action)
	if strings.Contains(action, "action=") {
		return true
	}
	return strings.HasPrefix(action, "/web/") && !strings.HasPrefix(action, "/web?")
}

func filterSidebarLinks(links []parser.MenuItem) []parser.MenuItem {
	if 0 == len(links) {
		return links
	}
	out := links[:0]
	for _, mi := range links {
		if settingsNavExcludedSection(mi.Name) || settingsNavExcludedLink(mi.Name, mi.Action) {
			continue
		}
		out = append(out, mi)
	}
	return out
}

func menuItemHasAllowedChild(parentID string, allMenus []parser.MenuItem, menuAllowed func(parser.MenuItem) bool) bool {
	for _, sub := range allMenus {
		if sub.ParentID == parentID && menuAllowed(sub) {
			return true
		}
	}
	return false
}

// collectSidebarLinks returns depth-first navigable descendants of parentID (not parent itself).
func collectSidebarLinks(parentID string, allMenus []parser.MenuItem, menuAllowed func(parser.MenuItem) bool) []parser.MenuItem {
	var children []parser.MenuItem
	for _, sub := range allMenus {
		if sub.ParentID == parentID && menuAllowed(sub) {
			children = append(children, sub)
		}
	}
	sort.Slice(children, func(i, j int) bool {
		if children[i].Sequence != children[j].Sequence {
			return children[i].Sequence < children[j].Sequence
		}
		return children[i].Name < children[j].Name
	})

	var out []parser.MenuItem
	for _, child := range children {
		if settingsNavExcludedSection(child.Name) || settingsNavExcludedLink(child.Name, child.Action) {
			continue
		}
		hasChildren := menuItemHasAllowedChild(child.ID, allMenus, menuAllowed)
		switch {
		case menuItemHasWindowOrURLAction(child):
			out = append(out, child)
		case !hasChildren:
			out = append(out, child)
		}
		if hasChildren {
			out = append(out, collectSidebarLinks(child.ID, allMenus, menuAllowed)...)
		}
	}
	return out
}

func sortMenuItemsBySequenceName(items []parser.MenuItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Sequence != items[j].Sequence {
			return items[i].Sequence < items[j].Sequence
		}
		return items[i].Name < items[j].Name
	})
}

func buildSidebarMenus(allMenus []parser.MenuItem, activeModuleID string, menuAllowed func(parser.MenuItem) bool) []SidebarMenu {
	if activeModuleID == "" {
		return nil
	}
	var sidebarMenus []SidebarMenu
	var sections []SidebarMenu
	for _, m := range allMenus {
		if m.ParentID != activeModuleID || !menuAllowed(m) || settingsNavExcludedSection(m.Name) {
			continue
		}
		section := SidebarMenu{ID: m.ID, Name: m.Name, Sequence: m.Sequence}
		links := collectSidebarLinks(m.ID, allMenus, menuAllowed)
		if menuItemHasWindowOrURLAction(m) && !settingsNavExcludedLink(m.Name, m.Action) {
			links = append([]parser.MenuItem{m}, links...)
		}
		sortMenuItemsBySequenceName(links)
		section.SubMenus = filterSidebarLinks(links)
		if 0 == len(section.SubMenus) {
			continue
		}
		sections = append(sections, section)
	}
	sort.Slice(sections, func(i, j int) bool {
		if sections[i].Sequence != sections[j].Sequence {
			return sections[i].Sequence < sections[j].Sequence
		}
		return sections[i].Name < sections[j].Name
	})
	sidebarMenus = append(sidebarMenus, sections...)
	return sidebarMenus
}

// SidebarHasMenus reports whether any sidebar section has at least one link.
func SidebarHasMenus(menus []SidebarMenu) bool {
	for _, sec := range menus {
		if len(sec.SubMenus) > 0 {
			return true
		}
	}
	return false
}

// sortTopBarRootMenus orders application root menus by each menuitem's `sequence` (then `name`),
// matching Sumeru XML.
func sortTopBarRootMenus(in []parser.MenuItem) []parser.MenuItem {
	if 0 == len(in) {
		return in
	}
	out := append([]parser.MenuItem{}, in...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func queryInstalledApplicationNames(ctx context.Context) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	if orm.DB == nil {
		return out, fmt.Errorf("database not initialized")
	}
	q := `SELECT name FROM ` + orm.MustQuotedTableName("sys.module") + ` WHERE state = 'installed' AND active = true AND application = true`
	rows, err := orm.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out[strings.TrimSpace(n)] = struct{}{}
	}
	return out, rows.Err()
}

// IsMenuUnderSettingsRoot reports whether activeMenuID is the Settings root or one of its descendants.
func IsMenuUnderSettingsRoot(ctx context.Context, activeMenuID string) bool {
	id64, err := strconv.ParseInt(strings.TrimSpace(activeMenuID), 10, 64)
	if err != nil || 0 >= id64 || orm.DB == nil {
		return false
	}
	rootID, _, err := orm.ResolveXmlId(ctx, "base.menu_settings_root")
	if err != nil || 0 == rootID {
		return false
	}
	return orm.MenuHasAncestor(ctx, int(id64), rootID)
}

type appLauncherItem struct {
	Kind        string `json:"kind,omitempty"` // "app" (default) or "menu"
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Href        string `json:"href"`
	MenuID      string `json:"menuId"`
	IconLetter  string `json:"iconLetter"`
}

// BuildAppLauncherJSON returns installed application metadata for the shell app launcher.
func BuildAppLauncherJSON(ctx context.Context) template.JS {
	if orm.DB == nil {
		return template.JS("[]")
	}
	modTbl := orm.MustQuotedTableName("sys.module")
	q := `SELECT name, COALESCE(NULLIF(TRIM(display_name), ''), name), COALESCE(description, '')
		FROM ` + modTbl + ` WHERE state = 'installed' AND active = true AND application = true ORDER BY display_name, name`
	rows, err := orm.DB.QueryContext(ctx, q)
	if err != nil {
		return template.JS("[]")
	}
	defer rows.Close()
	var items []appLauncherItem
	for rows.Next() {
		var name, displayName, desc string
		if err := rows.Scan(&name, &displayName, &desc); err != nil {
			continue
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		href := ModuleWorkspaceEntryURL(ctx, name)
		menuID := RootMenuIDForModule(ctx, name)
		items = append(items, appLauncherItem{
			Kind:        "app",
			Name:        name,
			DisplayName: strings.TrimSpace(displayName),
			Description: strings.TrimSpace(desc),
			Href:        href,
			MenuID:      fmt.Sprintf("%d", menuID),
			IconLetter:  IconLetterFromName(displayName),
		})
	}

	uid := orm.UIDFromContext(ctx)
	allMenus := fetchShellMenus(ctx)
	menuAllowed := func(mi parser.MenuItem) bool { return shellMenuAllowed(ctx, uid, mi) }
	for _, m := range allMenus {
		if !menuAllowed(m) {
			continue
		}
		if !strings.Contains(m.Action, "action=") {
			continue
		}
		modLabel := strings.TrimSpace(m.Module)
		if modLabel == "" {
			modLabel = "Menu"
		}
		items = append(items, appLauncherItem{
			Kind:        "menu",
			Name:        m.ID,
			DisplayName: strings.TrimSpace(m.Name),
			Description: modLabel,
			Href:        m.Action,
			MenuID:      m.ID,
			IconLetter:  IconLetterFromName(m.Name),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		ki, kj := items[i].Kind, items[j].Kind
		if ki != kj {
			if ki == "app" {
				return true
			}
			if kj == "app" {
				return false
			}
		}
		di, dj := strings.ToLower(items[i].DisplayName), strings.ToLower(items[j].DisplayName)
		if di != dj {
			return di < dj
		}
		return items[i].Name < items[j].Name
	})
	b, err := json.Marshal(items)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(b)
}
