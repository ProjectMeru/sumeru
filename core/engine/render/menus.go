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
	return orm.UserMayAccessMenu(ctx, uid, mi.AccessGroups)
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
	if len(allMenus) == 0 {
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
		  WHERE COALESCE(NULLIF(TRIM(m.module), ''), '') = ''
		     OR EXISTS (
		      SELECT 1 FROM %s im WHERE im.name = m.module AND im.state = 'installed' AND im.active = true
		    )
		  ORDER BY m.sequence`,
		menuTbl, modTbl,
	)
	rows, err := orm.DB.QueryContext(ctx, query)
	if err != nil {
		applog.WarnMsg(ctx, "render", "menus", "Error fetching menus", err, nil)
		return nil
	}
	defer rows.Close()

	lang := "en_US"
	if uid := orm.UIDFromContext(ctx); uid > 0 {
		if u, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": uid}); err == nil {
			if l := strings.TrimSpace(orm.AsString(u["lang"])); l != "" {
				lang = l
			}
		}
	}

	var allMenus []parser.MenuItem
	for rows.Next() {
		var id int
		var name, mod, webIcon, accessGroups string
		var parentID, actionID, seq sql.NullInt64
		if err := rows.Scan(&id, &name, &parentID, &actionID, &seq, &mod, &webIcon, &accessGroups); err != nil {
			applog.WarnMsg(ctx, "render", "menus", "Error scanning menu row", err, nil)
			continue
		}

		m := parser.MenuItem{
			ID:           fmt.Sprintf("%d", id),
			Name:         orm.Translate(ctx, lang, name),
			Sequence:     int(seq.Int64),
			Module:       mod,
			WebIcon:      sanitizeMenuIcon(webIcon),
			AccessGroups: strings.TrimSpace(accessGroups),
			Action:       fmt.Sprintf("/web?menu_id=%d", id),
		}
		if parentID.Valid && parentID.Int64 != int64(id) {
			m.ParentID = fmt.Sprintf("%d", parentID.Int64)
		}
		if actionID.Valid && actionID.Int64 != 0 {
			m.Action = fmt.Sprintf("/web?action=%d&menu_id=%d", actionID.Int64, id)
		} else if !parentID.Valid && strings.EqualFold(strings.TrimSpace(name), "Home") {
			m.Action = "/web/home"
		}
		allMenus = append(allMenus, m)
	}
	if err := rows.Err(); err != nil {
		applog.WarnMsg(ctx, "render", "menus", "Menu rows error", err, nil)
	}
	return allMenus
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

func buildSidebarMenus(allMenus []parser.MenuItem, activeModuleID string, menuAllowed func(parser.MenuItem) bool) []SidebarMenu {
	if activeModuleID == "" {
		return nil
	}
	var sidebarMenus []SidebarMenu
	var sections []SidebarMenu
	for _, m := range allMenus {
		if m.ParentID != activeModuleID || !menuAllowed(m) {
			continue
		}
		section := SidebarMenu{ID: m.ID, Name: m.Name, Sequence: m.Sequence}
		for _, sub := range allMenus {
			if sub.ParentID != m.ID || !menuAllowed(sub) {
				continue
			}
			section.SubMenus = append(section.SubMenus, sub)
		}
		sort.Slice(section.SubMenus, func(i, j int) bool {
			if section.SubMenus[i].Sequence != section.SubMenus[j].Sequence {
				return section.SubMenus[i].Sequence < section.SubMenus[j].Sequence
			}
			return section.SubMenus[i].Name < section.SubMenus[j].Name
		})
		if len(section.SubMenus) == 0 {
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
	if len(in) == 0 {
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
	if err != nil || id64 <= 0 || orm.DB == nil {
		return false
	}
	rootID, _, err := orm.ResolveXmlId(ctx, "base.menu_settings_root")
	if err != nil || rootID == 0 {
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
		menuID := RootMenuIDForModule(ctx, name)
		href := "/web/home"
		if menuID > 0 {
			href = fmt.Sprintf("/web?menu_id=%d", menuID)
		}
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
