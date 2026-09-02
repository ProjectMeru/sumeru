package web

import (
	"sort"
	"strings"

	"sumeru/core/module"
)

const appsCategoryUncategorized = "Uncategorized"

type appsModuleGroup struct {
	Label   string
	Modules []appsModule
}

type appsCategoryNavItem struct {
	Name   string
	Label  string
	Count  int
	Href   string
	Active bool
}

type appsCategoryNavVM struct {
	AllApps appsCategoryNavItem
	Items   []appsCategoryNavItem
}

type appsCatalogStats struct {
	VisibleApps      int
	InstalledApps    int
	BrowseTotalApps  int
	HasActiveFilters bool
}

func buildAppsCatalogStats(visible, browseScope []appsModule, browse appsBrowseState) appsCatalogStats {
	installed := 0
	for _, mod := range visible {
		if mod.State == moduleStateInstalled {
			installed++
		}
	}
	return appsCatalogStats{
		VisibleApps:     len(visible),
		InstalledApps:   installed,
		BrowseTotalApps: len(browseScope),
		HasActiveFilters: strings.TrimSpace(browse.SearchQuery) != "" ||
			strings.TrimSpace(browse.Category) != "" ||
			browse.Filter != appsFilterAll ||
			browse.Scope != appsScopeAll ||
			browse.GroupBy != "",
	}
}

func moduleSummary(name, description string) string {
	if addon, ok := module.DiscoveredAddons[name]; ok {
		if summary := strings.TrimSpace(addon.Manifest.Summary); summary != "" {
			return summary
		}
	}
	return moduleSummaryFromDescription(description)
}

func moduleSummaryFromDescription(description string) string {
	description = strings.TrimSpace(description)
	if description == "" {
		return ""
	}
	if idx := strings.IndexAny(description, "\n\r"); idx >= 0 {
		return strings.TrimSpace(description[:idx])
	}
	if len(description) > 160 {
		return strings.TrimSpace(description[:157]) + "..."
	}
	return description
}

func moduleHasLongDescription(summary, description string) bool {
	description = strings.TrimSpace(description)
	if description == "" {
		return false
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return true
	}
	return strings.TrimSpace(description) != summary && len(description) > len(summary)+20
}

func moduleCategoryLabel(categoryName string) string {
	categoryName = strings.TrimSpace(categoryName)
	if categoryName == "" {
		return appsCategoryUncategorized
	}
	return categoryName
}

func appsModuleMatchesCategory(moduleEntry appsModule, categoryFilter string) bool {
	categoryFilter = strings.TrimSpace(categoryFilter)
	if categoryFilter == "" {
		return true
	}
	return strings.EqualFold(moduleCategoryLabel(moduleEntry.Category), categoryFilter)
}

func groupAppsModules(modules []appsModule, groupBy string) []appsModuleGroup {
	if groupBy != appsGroupByCategory {
		if len(modules) == 0 {
			return nil
		}
		return []appsModuleGroup{{Modules: modules}}
	}
	byLabel := map[string][]appsModule{}
	for _, mod := range modules {
		label := moduleCategoryLabel(mod.Category)
		byLabel[label] = append(byLabel[label], mod)
	}
	labels := make([]string, 0, len(byLabel))
	for label := range byLabel {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	if idx := sort.SearchStrings(labels, appsCategoryUncategorized); idx < len(labels) && labels[idx] == appsCategoryUncategorized {
		labels = append(append(labels[:idx], labels[idx+1:]...), appsCategoryUncategorized)
	}
	out := make([]appsModuleGroup, 0, len(labels))
	for _, label := range labels {
		group := byLabel[label]
		sort.Slice(group, func(i, j int) bool {
			return strings.ToLower(group[i].DisplayName) < strings.ToLower(group[j].DisplayName)
		})
		out = append(out, appsModuleGroup{Label: label, Modules: group})
	}
	return out
}

func buildAppsCategoryNav(modules []appsModule, browse appsBrowseState) appsCategoryNavVM {
	counts := map[string]int{}
	for _, mod := range modules {
		if !mod.Application {
			continue
		}
		counts[moduleCategoryLabel(mod.Category)]++
	}
	allBrowse := browse
	allBrowse.Category = ""
	items := make([]appsCategoryNavItem, 0, len(counts))
	for label, count := range counts {
		itemBrowse := browse
		if label != appsCategoryUncategorized {
			itemBrowse.Category = label
		} else {
			itemBrowse.Category = appsCategoryUncategorized
		}
		items = append(items, appsCategoryNavItem{
			Name:   label,
			Label:  label,
			Count:  count,
			Href:   appsLinkFromBrowse(itemBrowse),
			Active: strings.EqualFold(browse.Category, label),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Label) < strings.ToLower(items[j].Label)
	})
	total := 0
	for _, mod := range modules {
		if mod.Application {
			total++
		}
	}
	return appsCategoryNavVM{
		AllApps: appsCategoryNavItem{
			Label:  "All applications",
			Count:  total,
			Href:   appsLinkFromBrowse(allBrowse),
			Active: strings.TrimSpace(browse.Category) == "",
		},
		Items: items,
	}
}
