package render

import (
	"net/url"
	"strings"
)

func appendAppsTabQuery(q url.Values, layout, msg, module, filter, scope, search, category, groupBy string) {
	if layout == "list" || layout == "grid" {
		q.Set("layout", layout)
	}
	if strings.TrimSpace(msg) != "" {
		q.Set("msg", strings.TrimSpace(msg))
	}
	if strings.TrimSpace(module) != "" {
		q.Set("module", strings.TrimSpace(module))
	}
	if filter != "" && filter != "all" {
		q.Set("filter", filter)
	}
	if scope != "" && scope != "all" {
		q.Set("scope", scope)
	}
	if strings.TrimSpace(search) != "" {
		q.Set("q", strings.TrimSpace(search))
	}
	if strings.TrimSpace(category) != "" {
		q.Set("category", strings.TrimSpace(category))
	}
	if strings.TrimSpace(groupBy) != "" {
		q.Set("group_by", strings.TrimSpace(groupBy))
	}
}

// AppsViewTabs builds Grid / List links for the Apps dashboard (?layout=) preserving browse params.
func AppsViewTabs(currentLayout, msg, module, filter, scope, search, category, groupBy string) []ViewSwitchTab {
	cur := strings.ToLower(strings.TrimSpace(currentLayout))
	if cur == "" {
		cur = "grid"
	}
	if cur == "kanban" {
		cur = "grid"
	}
	msg = strings.TrimSpace(msg)
	module = strings.TrimSpace(module)
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" {
		filter = "all"
	}
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		scope = "all"
	}
	search = strings.TrimSpace(search)
	category = strings.TrimSpace(category)
	groupBy = strings.TrimSpace(groupBy)

	order := []struct {
		layoutKey string
		label     string
		mode      string
	}{
		{"grid", "Grid", "apps_grid"},
		{"list", "List", "apps_list"},
	}
	var out []ViewSwitchTab
	for _, o := range order {
		q := url.Values{}
		appendAppsTabQuery(q, o.layoutKey, msg, module, filter, scope, search, category, groupBy)
		out = append(out, ViewSwitchTab{
			Label:  o.label,
			Href:   "/web/apps?" + q.Encode(),
			Mode:   o.mode,
			Active: cur == o.layoutKey,
		})
	}
	return out
}

// HomeViewTabs builds Grid / List links for the Home dashboard.
func HomeViewTabs(currentLayout string) []ViewSwitchTab {
	cur := strings.ToLower(strings.TrimSpace(currentLayout))
	if cur == "" {
		cur = "grid"
	}
	if cur == "kanban" {
		cur = "grid"
	}
	if cur != "list" {
		cur = "grid"
	}

	order := []struct {
		layoutKey string
		label     string
		mode      string
	}{
		{"grid", "Grid", "apps_grid"},
		{"list", "List", "apps_list"},
	}
	var out []ViewSwitchTab
	for _, o := range order {
		q := url.Values{}
		if o.layoutKey == "list" || o.layoutKey == "grid" {
			q.Set("layout", o.layoutKey)
		}
		out = append(out, ViewSwitchTab{
			Label:  o.label,
			Href:   "/web/home?" + q.Encode(),
			Mode:   o.mode,
			Active: cur == o.layoutKey,
		})
	}
	return out
}
