package web

import (
	"net/http"
	"net/url"
	"strings"
)

type appsNavVM struct {
	FilterAll, FilterInstalled, FilterUninstalled string
	ScopeAll, ScopeApps, ScopeTechnical           string
}

// appsBrowseState holds normalized Apps page query parameters.
type appsBrowseState struct {
	Message     string
	Layout      string
	Filter      string
	Scope       string
	SearchQuery string
	Category    string
	GroupBy     string
	ModuleName  string
	Editing     bool
}

func parseAppsBrowseState(r *http.Request) appsBrowseState {
	query := r.URL.Query()
	return appsBrowseState{
		Message:     strings.TrimSpace(query.Get("msg")),
		Layout:      layoutFromQuery(r),
		ModuleName:  strings.TrimSpace(query.Get("module")),
		Editing:     strings.TrimSpace(query.Get("edit")) == "1",
		Filter:      normalizeAppsChoiceParam("filter", query.Get("filter")),
		Scope:       normalizeAppsChoiceParam("scope", query.Get("scope")),
		SearchQuery: strings.TrimSpace(query.Get("q")),
		Category:    strings.TrimSpace(query.Get("category")),
		GroupBy:     normalizeAppsChoiceParam("group_by", query.Get("group_by")),
	}
}

// parseAppsBrowseStateFromForm reads apps_* POST fields preserved across module action redirects.
func parseAppsBrowseStateFromForm(r *http.Request) appsBrowseState {
	return appsBrowseState{
		Layout:      layoutFromForm(r, appsLayoutField),
		Filter:      normalizeAppsChoiceParam("filter", r.FormValue(appsFilterField)),
		Scope:       normalizeAppsChoiceParam("scope", r.FormValue(appsScopeField)),
		SearchQuery: strings.TrimSpace(r.FormValue(appsSearchField)),
		Category:    strings.TrimSpace(r.FormValue(appsCategoryField)),
		GroupBy:     normalizeAppsChoiceParam("group_by", r.FormValue(appsGroupByField)),
	}
}

// appsRedirectURL builds an Apps page URL (list or module detail from browse.ModuleName).
func appsRedirectURL(message string, browse appsBrowseState) string {
	return appsPageURL(message, browse)
}

func appsPageURL(message string, browse appsBrowseState) string {
	query := url.Values{}
	if trimmed := strings.TrimSpace(message); trimmed != "" {
		query.Set("msg", trimmed)
	}
	appendAppsBrowseQuery(query, browse)
	return buildAppsPageURL(query)
}

func appsLinkFromBrowse(browse appsBrowseState) string {
	return appsPageURL("", browse)
}

func appsDetailPageURL(browse appsBrowseState, editing bool) string {
	query := url.Values{}
	appendAppsBrowseQuery(query, browse)
	if editing {
		query.Set("edit", "1")
	}
	return buildAppsPageURL(query)
}

func buildAppsPageURL(query url.Values) string {
	if encoded := query.Encode(); encoded != "" {
		return appsRoute + "?" + encoded
	}
	return appsRoute
}

func withModuleName(browse appsBrowseState, moduleName string) appsBrowseState {
	browse.ModuleName = moduleName
	return browse
}

func appendAppsBrowseQuery(query url.Values, browse appsBrowseState) {
	appendAppsQueryBase(query, browse)
	if browse.ModuleName != "" {
		query.Set("module", browse.ModuleName)
	}
}

func normalizeAppsChoiceParam(kind, raw string) string {
	allowed, ok := appsAllowedChoices[kind]
	if !ok {
		return strings.ToLower(strings.TrimSpace(raw))
	}
	return normalizeAppsChoice(raw, allowed...)
}

var appsAllowedChoices = map[string][]string{
	"filter":   {appsFilterInstalled, appsFilterUninstalled, appsFilterAll},
	"scope":    {appsScopeApps, appsScopeTechnical, appsScopeAll},
	"group_by": {appsGroupByCategory, ""},
}

func normalizeAppsChoice(raw string, allowed ...string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	for _, choice := range allowed {
		if normalized == choice {
			return choice
		}
	}
	if len(allowed) > 0 {
		return allowed[len(allowed)-1]
	}
	return normalized
}

func appsModuleMatchesFilter(moduleEntry appsModule, filter string) bool {
	switch filter {
	case appsFilterInstalled:
		return moduleEntry.State == moduleStateInstalled
	case appsFilterUninstalled:
		return moduleEntry.State != moduleStateInstalled
	default:
		return true
	}
}

func appsModuleMatchesSearch(moduleEntry appsModule, searchQuery string) bool {
	searchQuery = strings.TrimSpace(searchQuery)
	if searchQuery == "" {
		return true
	}
	needle := strings.ToLower(searchQuery)
	searchText := strings.ToLower(moduleEntry.Name + " " + moduleEntry.DisplayName)
	return strings.Contains(searchText, needle)
}

// filterAppsModulesByBrowse applies search/filter/category and splits by application vs technical scope.
func filterAppsModulesByBrowse(modules []appsModule, browse appsBrowseState) (appModules, techModules []appsModule) {
	for _, moduleEntry := range modules {
		if !appsModuleMatchesSearch(moduleEntry, browse.SearchQuery) ||
			!appsModuleMatchesFilter(moduleEntry, browse.Filter) ||
			!appsModuleMatchesCategory(moduleEntry, browse.Category) {
			continue
		}
		if moduleEntry.Application {
			appModules = append(appModules, moduleEntry)
		} else {
			techModules = append(techModules, moduleEntry)
		}
	}
	switch browse.Scope {
	case appsScopeApps:
		techModules = nil
	case appsScopeTechnical:
		appModules = nil
	}
	return appModules, techModules
}

func appendAppsQueryBase(query url.Values, browse appsBrowseState) {
	if browse.Layout == appsLayoutList || browse.Layout == appsLayoutGrid {
		query.Set("layout", browse.Layout)
	}
	if browse.Filter != "" && browse.Filter != appsFilterAll {
		query.Set("filter", browse.Filter)
	}
	if browse.Scope != "" && browse.Scope != appsScopeAll {
		query.Set("scope", browse.Scope)
	}
	if trimmed := strings.TrimSpace(browse.SearchQuery); trimmed != "" {
		query.Set("q", trimmed)
	}
	if trimmed := strings.TrimSpace(browse.Category); trimmed != "" {
		query.Set("category", trimmed)
	}
	if browse.GroupBy == appsGroupByCategory {
		query.Set("group_by", browse.GroupBy)
	}
}

func appsBrowseQuery(browse appsBrowseState) string {
	query := url.Values{}
	appendAppsBrowseQuery(query, browse)
	return query.Encode()
}

// appsDetailURL builds links for module detail, edit, and cancel views.
func appsDetailURL(browse appsBrowseState, editing bool) string {
	return appsDetailPageURL(browse, editing)
}

func buildAppsNavVM(browse appsBrowseState) appsNavVM {
	withFilter := func(filter string) string {
		linkBrowse := browse
		linkBrowse.Filter = filter
		return appsLinkFromBrowse(linkBrowse)
	}
	withScope := func(scope string) string {
		linkBrowse := browse
		linkBrowse.Scope = scope
		return appsLinkFromBrowse(linkBrowse)
	}
	return appsNavVM{
		FilterAll:         withFilter(appsFilterAll),
		FilterInstalled:   withFilter(appsFilterInstalled),
		FilterUninstalled: withFilter(appsFilterUninstalled),
		ScopeAll:          withScope(appsScopeAll),
		ScopeApps:         withScope(appsScopeApps),
		ScopeTechnical:    withScope(appsScopeTechnical),
	}
}
