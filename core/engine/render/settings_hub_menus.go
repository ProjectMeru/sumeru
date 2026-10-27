package render

import (
	"context"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// SettingsHubLink is one navigable entry on the settings hub.
type SettingsHubLink struct {
	Name      string
	Href      string
	MenuID    string
	WebIcon   string
	ActionKey string
}

// SettingsHubGroup is an optional subgroup inside a category card.
type SettingsHubGroup struct {
	Title string
	Links []SettingsHubLink
}

// SettingsHubCategory is a top-level settings hub section (settings root child).
type SettingsHubCategory struct {
	Title      string
	Sequence   int
	FilterText string
	SingleLink bool
	Groups     []SettingsHubGroup
}

func hubLinkActionKey(mi parser.MenuItem) string {
	action := strings.TrimSpace(mi.Action)
	if idx := strings.Index(action, "action="); idx >= 0 {
		rest := action[idx+len("action="):]
		if amp := strings.Index(rest, "&"); amp >= 0 {
			rest = rest[:amp]
		}
		rest = strings.TrimSpace(rest)
		if rest != "" {
			return "action:" + rest
		}
	}
	return "menu:" + strings.TrimSpace(mi.ID)
}

func hubLinkFromMenu(mi parser.MenuItem, seen map[string]struct{}) (SettingsHubLink, bool) {
	name := strings.TrimSpace(mi.Name)
	href := strings.TrimSpace(mi.Action)
	if name == "" || href == "" || settingsNavExcludedLink(name, href) {
		return SettingsHubLink{}, false
	}
	key := hubLinkActionKey(mi)
	if _, dup := seen[key]; dup {
		return SettingsHubLink{}, false
	}
	seen[key] = struct{}{}
	return SettingsHubLink{
		Name:      name,
		Href:      href,
		MenuID:    mi.ID,
		WebIcon:   mi.WebIcon,
		ActionKey: key,
	}, true
}

func settingsHubFilterText(categoryTitle string, groups []SettingsHubGroup) string {
	terms := []string{strings.ToLower(strings.TrimSpace(categoryTitle))}
	for _, g := range groups {
		if t := strings.TrimSpace(g.Title); t != "" {
			terms = append(terms, strings.ToLower(t))
		}
		for _, l := range g.Links {
			if n := strings.TrimSpace(l.Name); n != "" {
				terms = append(terms, strings.ToLower(n))
			}
		}
	}
	return strings.Join(terms, " ")
}

func buildSettingsHubCategory(catNode parser.MenuItem, allMenus []parser.MenuItem, menuAllowed func(parser.MenuItem) bool, seen map[string]struct{}) SettingsHubCategory {
	var groups []SettingsHubGroup
	var defaultLinks []SettingsHubLink

	if menuItemHasWindowOrURLAction(catNode) {
		if link, ok := hubLinkFromMenu(catNode, seen); ok {
			defaultLinks = append(defaultLinks, link)
		}
	}

	var children []parser.MenuItem
	for _, sub := range allMenus {
		if sub.ParentID == catNode.ID && menuAllowed(sub) {
			children = append(children, sub)
		}
	}
	sortMenuItemsBySequenceName(children)

	for _, child := range children {
		if settingsNavExcludedSection(child.Name) {
			continue
		}
		hasChildren := menuItemHasAllowedChild(child.ID, allMenus, menuAllowed)
		isContainer := hasChildren && !menuItemHasWindowOrURLAction(child)
		if isContainer {
			subMenus := collectSidebarLinks(child.ID, allMenus, menuAllowed)
			var links []SettingsHubLink
			for _, mi := range subMenus {
				if link, ok := hubLinkFromMenu(mi, seen); ok {
					links = append(links, link)
				}
			}
			if len(links) > 0 {
				groups = append(groups, SettingsHubGroup{Title: child.Name, Links: links})
			}
			continue
		}
		if menuItemHasWindowOrURLAction(child) || !hasChildren {
			if link, ok := hubLinkFromMenu(child, seen); ok {
				defaultLinks = append(defaultLinks, link)
			}
		}
	}

	if len(defaultLinks) > 0 {
		groups = append([]SettingsHubGroup{{Links: defaultLinks}}, groups...)
	}

	singleLink := len(groups) == 1 && len(groups[0].Links) == 1 && groups[0].Title == "" &&
		strings.EqualFold(strings.TrimSpace(groups[0].Links[0].Name), strings.TrimSpace(catNode.Name))

	return SettingsHubCategory{
		Title:      catNode.Name,
		Sequence:   catNode.Sequence,
		Groups:     groups,
		SingleLink: singleLink,
		FilterText: settingsHubFilterText(catNode.Name, groups),
	}
}

// BuildSettingsHubCategories lists settings hub cards from the menu tree under settingsRootMenuID.
func BuildSettingsHubCategories(ctx context.Context, settingsRootMenuID string) []SettingsHubCategory {
	settingsRootMenuID = strings.TrimSpace(settingsRootMenuID)
	if settingsRootMenuID == "" {
		return nil
	}
	allMenus := fetchShellMenus(ctx)
	if 0 == len(allMenus) {
		return nil
	}
	uid := orm.UIDFromContext(ctx)
	menuAllowed := func(mi parser.MenuItem) bool { return shellMenuAllowed(ctx, uid, mi) }

	var rootChildren []parser.MenuItem
	for _, m := range allMenus {
		if m.ParentID == settingsRootMenuID && menuAllowed(m) {
			rootChildren = append(rootChildren, m)
		}
	}
	sortMenuItemsBySequenceName(rootChildren)

	seen := make(map[string]struct{})
	var categories []SettingsHubCategory
	for _, catNode := range rootChildren {
		if isSettingsHubNavMenu(ctx, catNode.ID) || settingsNavExcludedSection(catNode.Name) {
			continue
		}
		cat := buildSettingsHubCategory(catNode, allMenus, menuAllowed, seen)
		if 0 == len(cat.Groups) {
			continue
		}
		categories = append(categories, cat)
	}
	return categories
}

// BuildSettingsHubCategoriesForTest builds hub categories from an in-memory menu tree (tests).
func BuildSettingsHubCategoriesForTest(allMenus []parser.MenuItem, settingsRootMenuID string, menuAllowed func(parser.MenuItem) bool, skipMenuIDs ...string) []SettingsHubCategory {
	settingsRootMenuID = strings.TrimSpace(settingsRootMenuID)
	if settingsRootMenuID == "" || menuAllowed == nil {
		return nil
	}
	skip := map[string]struct{}{}
	for _, id := range skipMenuIDs {
		if id = strings.TrimSpace(id); id != "" {
			skip[id] = struct{}{}
		}
	}
	var rootChildren []parser.MenuItem
	for _, m := range allMenus {
		if m.ParentID == settingsRootMenuID && menuAllowed(m) {
			rootChildren = append(rootChildren, m)
		}
	}
	sortMenuItemsBySequenceName(rootChildren)
	seen := make(map[string]struct{})
	var categories []SettingsHubCategory
	for _, catNode := range rootChildren {
		if _, omit := skip[catNode.ID]; omit {
			continue
		}
		if settingsNavExcludedSection(catNode.Name) {
			continue
		}
		cat := buildSettingsHubCategory(catNode, allMenus, menuAllowed, seen)
		if 0 == len(cat.Groups) {
			continue
		}
		categories = append(categories, cat)
	}
	return categories
}
