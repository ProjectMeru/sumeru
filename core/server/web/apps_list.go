package web

import (
	"context"
	"net/http"

	"sumeru/core/engine/render"
)

type appsModuleDep struct {
	Name        string
	DisplayName string
	DetailURL   string
}

type appsModule struct {
	Name          string
	DisplayName   string
	Author        string
	Version       string
	Description   string
	Summary       string
	Category      string
	State         string
	Application   bool
	Active        bool
	IsCore        bool
	CanInstall    bool
	CanUninstall  bool
	CanDeactivate bool
	CanActivate   bool
	IconLetter    string
	IconURL       string
	IconHue       int
	DetailURL     string
}

type appsPageData struct {
	Title           string
	CSRFToken       string
	Modules         []appsModule
	AppModules      []appsModule
	TechModules     []appsModule
	AppGroups       []appsModuleGroup
	Layout          string
	Filter          string
	Scope           string
	Search          string
	Category        string
	GroupBy         string
	ShowCategoryCol bool
	CategoryNav     appsCategoryNavVM
	Stats           appsCatalogStats
	Nav             appsNavVM
	ModuleDetail    *appsModuleDetailVM
	ViewBreadcrumb  string
}

const appsPageScriptURL = "/static/js/apps-page.js"

// AppsHandler lists installable apps and exposes install / uninstall / activate controls.
func AppsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSystemAdmin(w, r, true) {
		return
	}

	ctx := r.Context()
	browse := parseAppsBrowseState(r)

	moduleRows, ok := listModulesOr500(w, r, "Failed to list modules for apps page")
	if !ok {
		return
	}

	allModules := enrichAppsModules(ctx, buildAppsModuleList(moduleRows), browse)
	navBrowse := browse
	navBrowse.Category = ""
	navAppModules, _ := filterAppsModulesByBrowse(allModules, navBrowse)
	appModules, techModules := filterAppsModulesByBrowse(allModules, browse)
	appGroups := groupAppsModules(appModules, browse.GroupBy)
	categoryNav := buildAppsCategoryNav(navAppModules, browse)
	catalogStats := buildAppsCatalogStats(appModules, navAppModules, browse)

	detail, breadcrumb, ok := loadAppsModuleDetail(ctx, w, r, browse, allModules)
	if !ok {
		return
	}

	listHref := appsLinkFromBrowse(browse)
	detailTitle := ""
	if detail != nil {
		detailTitle = detail.DisplayName
	}

	inlineFlashes, toastFlashes := splitAppsPageFlashes(browse.Message, buildModuleDisplayNameMap(allModules))

	page := render.PageData{
		Title:           appsPageTitle,
		ViewBreadcrumb:  breadcrumb,
		ModuleName:      appsPageTitle,
		AppsNavActive:   true,
		SuppressSidebar: true,
		FlashMessages:   inlineFlashes,
		ToastMessages:   toastFlashes,
		ExtraScriptURLs: []string{appsPageScriptURL},
		ViewTabs: render.AppsViewTabs(
			browse.Layout,
			browse.Message,
			browse.ModuleName,
			browse.Filter,
			browse.Scope,
			browse.SearchQuery,
			browse.Category,
			browse.GroupBy,
		),
		BreadcrumbItems: render.BuildAppsBreadcrumbs(ctx, listHref, detailTitle),
	}
	if detail != nil {
		page.ActivityContextModel = appsModuleModel
		page.ActivityContextRecordID = int64(detail.ID)
	}

	renderShellPage(w, r, shellPageOpts{
		Route:         appsRoute,
		InnerTemplate: appsInnerTemplate,
		InnerData: appsPageData{
			Title:           appsPageTitle,
			CSRFToken:       CSRFTokenForRequest(r),
			Modules:         allModules,
			AppModules:      appModules,
			TechModules:     techModules,
			AppGroups:       appGroups,
			Layout:          browse.Layout,
			Filter:          browse.Filter,
			Scope:           browse.Scope,
			Search:          browse.SearchQuery,
			Category:        browse.Category,
			GroupBy:         browse.GroupBy,
			ShowCategoryCol: browse.GroupBy != appsGroupByCategory,
			CategoryNav:     categoryNav,
			Stats:           catalogStats,
			Nav:             buildAppsNavVM(browse),
			ModuleDetail:    detail,
			ViewBreadcrumb:  breadcrumb,
		},
		Page: page,
	})

	logAppsPageOpen(ctx, r.URL.Path, browse)
}

func buildAppsModuleList(moduleRows []map[string]interface{}) []appsModule {
	modules := make([]appsModule, 0, len(moduleRows))
	for _, row := range moduleRows {
		parsed, rowOK := parseModuleRow(row)
		if !rowOK {
			continue
		}
		modules = append(modules, appsModuleFromParsed(parsed))
	}
	return modules
}

func enrichAppsModules(ctx context.Context, modules []appsModule, browse appsBrowseState) []appsModule {
	if len(modules) == 0 {
		return modules
	}
	out := make([]appsModule, len(modules))
	for i, mod := range modules {
		out[i] = mod
		out[i].IconURL = render.ModuleIconURL(ctx, mod.Name)
		out[i].IconHue = render.IconHueFromString(mod.Name)
		out[i].DetailURL = appsDetailPageURL(withModuleName(browse, mod.Name), false)
	}
	return out
}

// appsModuleFromParsed maps a normalized module row to the Apps list view model, including action flags.
func appsModuleFromParsed(parsed moduleRow) appsModule {
	isCore := parsed.Name == "base"
	isInstalled := parsed.State == "installed"
	summary := moduleSummary(parsed.Name, parsed.Description)
	return appsModule{
		Name:          parsed.Name,
		DisplayName:   parsed.DisplayName,
		Author:        parsed.Author,
		Version:       parsed.Version,
		Description:   parsed.Description,
		Summary:       summary,
		Category:      moduleCategoryLabel(parsed.CategoryName),
		State:         parsed.State,
		Application:   parsed.Application,
		Active:        parsed.Active,
		IsCore:        isCore,
		CanInstall:    !isInstalled,
		CanUninstall:  isInstalled && !isCore,
		CanDeactivate: isInstalled && parsed.Active && !isCore,
		CanActivate:   isInstalled && !parsed.Active && !isCore,
		IconLetter:    render.IconLetterFromName(parsed.DisplayName),
	}
}

func logAppsPageOpen(ctx context.Context, route string, browse appsBrowseState) {
	fields := map[string]interface{}{
		"layout":   browse.Layout,
		"filter":   browse.Filter,
		"scope":    browse.Scope,
		"search":   browse.SearchQuery,
		"category": browse.Category,
		"group_by": browse.GroupBy,
	}
	if browse.ModuleName != "" {
		fields["module"] = browse.ModuleName
	}
	WebLogNavigation(ctx, route, "apps_open", "Apps page opened", fields)
}
