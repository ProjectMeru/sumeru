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

func normalizeAppsHubLayout(layout string, collapseNonList bool) string {
	cur := strings.ToLower(strings.TrimSpace(layout))
	if cur == "" {
		cur = "grid"
	}
	if cur == "kanban" {
		cur = "grid"
	}
	if collapseNonList && cur != "list" {
		cur = "grid"
	}
	return cur
}

var appsHubLayoutTabOrder = []struct {
	layoutKey string
	label     string
	mode      string
}{
	{"grid", "Grid", "apps_grid"},
	{"list", "List", "apps_list"},
}

func buildAppsHubLayoutTabs(basePath string, cur string, fill func(q url.Values, layoutKey string)) []ViewSwitchTab {
	out := make([]ViewSwitchTab, 0, len(appsHubLayoutTabOrder))
	for _, o := range appsHubLayoutTabOrder {
		q := url.Values{}
		fill(q, o.layoutKey)
		out = append(out, ViewSwitchTab{
			Label:  o.label,
			Href:   basePath + "?" + q.Encode(),
			Mode:   o.mode,
			Active: cur == o.layoutKey,
		})
	}
	return out
}

// AppsViewTabs builds Grid / List links for the Apps dashboard (?layout=) preserving browse params.
func AppsViewTabs(currentLayout, msg, module, filter, scope, search, category, groupBy string) []ViewSwitchTab {
	cur := normalizeAppsHubLayout(currentLayout, false)
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

	return buildAppsHubLayoutTabs("/web/apps", cur, func(q url.Values, layoutKey string) {
		appendAppsTabQuery(q, layoutKey, msg, module, filter, scope, search, category, groupBy)
	})
}

// HomeViewTabs builds Grid / List links for the Home dashboard.
func HomeViewTabs(currentLayout string) []ViewSwitchTab {
	cur := normalizeAppsHubLayout(currentLayout, true)
	return buildAppsHubLayoutTabs("/web/home", cur, func(q url.Values, layoutKey string) {
		appendAppsTabQuery(q, layoutKey, "", "", "all", "all", "", "", "")
	})
}
