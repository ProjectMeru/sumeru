package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/module"
	"sumeru/core/orm"
)

// appsModuleDetailVM is the read-only app detail view for one sys.module row on the Apps screen.
type appsModuleDetailVM struct {
	Layout                             string
	Editing                            bool
	Name, DisplayName, Author, Version string
	Description, Summary, CategoryLabel string
	State, StatusLabel                 string
	HasLongDescription                 bool
	Active                             bool
	IsApplication, IsCore              bool
	ID                                 int
	IconURL, IconLetter                string
	IconHue                            int
	OpenAppURL                         string
	Depends                            []appsModuleDep
	CanInstall, CanUninstall           bool
	CanDeactivate, CanActivate         bool
	BackAppsQuery                      string
	EditURL, CancelURL                 string
}

const appsListBreadcrumb = "Applications"

// loadAppsModuleDetail builds the detail VM when ?module= is present, or nil for list-only view.
func loadAppsModuleDetail(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	browse appsBrowseState,
	listedModules []appsModule,
) (detail *appsModuleDetailVM, breadcrumb string, ok bool) {
	breadcrumb = appsListBreadcrumb
	if browse.ModuleName == "" {
		return nil, breadcrumb, true
	}

	parsed, recordID, loaded := loadModuleRowByName(w, r, browse.ModuleName)
	if !loaded {
		return nil, "", false
	}

	listEntry, found := findAppsModule(listedModules, parsed.Name)
	if !found {
		listEntry = appsModuleFromParsed(parsed)
		listEntry.IconURL = render.ModuleIconURL(ctx, parsed.Name)
		listEntry.IconHue = render.IconHueFromString(parsed.Name)
		listEntry.DetailURL = appsDetailPageURL(withModuleName(browse, parsed.Name), false)
	}

	moduleBrowse := withModuleName(browse, parsed.Name)
	browseQuery := appsBrowseQuery(browse)
	summary := moduleSummary(parsed.Name, parsed.Description)

	detail = &appsModuleDetailVM{
		Layout:             browse.Layout,
		Editing:            browse.Editing,
		Name:               parsed.Name,
		DisplayName:        parsed.DisplayName,
		Author:             parsed.Author,
		Version:            parsed.Version,
		Description:        parsed.Description,
		Summary:            summary,
		CategoryLabel:      moduleCategoryLabel(parsed.CategoryName),
		HasLongDescription: moduleHasLongDescription(summary, parsed.Description),
		State:              parsed.State,
		StatusLabel:        moduleStatusLabel(parsed),
		Active:        parsed.Active,
		IsApplication: parsed.Application,
		IsCore:        listEntry.IsCore,
		ID:            recordID,
		IconURL:       listEntry.IconURL,
		IconLetter:    listEntry.IconLetter,
		IconHue:       listEntry.IconHue,
		OpenAppURL:    moduleOpenAppURL(ctx, parsed),
		Depends:       moduleDepends(parsed.Name, listedModules, browse),
		CanInstall:    listEntry.CanInstall,
		CanUninstall:  listEntry.CanUninstall,
		CanDeactivate: listEntry.CanDeactivate,
		CanActivate:   listEntry.CanActivate,
		BackAppsQuery: browseQuery,
		EditURL:       appsDetailPageURL(moduleBrowse, true),
		CancelURL:     appsDetailPageURL(moduleBrowse, false),
	}
	breadcrumb = "Apps · " + detail.DisplayName
	return detail, breadcrumb, true
}

func moduleStatusLabel(parsed moduleRow) string {
	switch parsed.State {
	case "installed":
		if parsed.Active {
			return "Installed"
		}
		return "Installed (inactive)"
	default:
		return "Not installed"
	}
}

func moduleOpenAppURL(ctx context.Context, parsed moduleRow) string {
	if parsed.State != "installed" || !parsed.Active {
		return ""
	}
	menuID := render.RootMenuIDForModule(ctx, parsed.Name)
	if menuID <= 0 {
		return ""
	}
	return fmt.Sprintf("/web?menu_id=%d", menuID)
}

func moduleDepends(moduleName string, listedModules []appsModule, browse appsBrowseState) []appsModuleDep {
	addon, ok := module.DiscoveredAddons[moduleName]
	if !ok {
		return nil
	}
	deps := addon.Manifest.Depends
	if len(deps) == 0 {
		return nil
	}
	out := make([]appsModuleDep, 0, len(deps))
	for _, depName := range deps {
		depName = strings.TrimSpace(depName)
		if depName == "" {
			continue
		}
		out = append(out, appsModuleDep{
			Name:        depName,
			DisplayName: moduleDepDisplayName(depName, listedModules),
			DetailURL:   appsDetailPageURL(withModuleName(browse, depName), false),
		})
	}
	return out
}

func moduleDepDisplayName(depName string, listedModules []appsModule) string {
	if entry, ok := findAppsModule(listedModules, depName); ok {
		return entry.DisplayName
	}
	if addon, ok := module.DiscoveredAddons[depName]; ok {
		return moduleDisplayName(depName, addon.Manifest.DisplayName)
	}
	return depName
}

// loadModuleRowByName fetches one sys.module row by technical name.
func loadModuleRowByName(w http.ResponseWriter, r *http.Request, moduleName string) (moduleRow, int, bool) {
	row, err := orm.SearchOne(r.Context(), appsModuleModel, map[string]interface{}{"name": moduleName})
	if err != nil {
		http.Error(w, "Module not found", http.StatusNotFound)
		return moduleRow{}, 0, false
	}
	parsed, rowOK := parseModuleRow(row)
	if !rowOK {
		http.Error(w, "Module not found", http.StatusNotFound)
		return moduleRow{}, 0, false
	}
	recordID, idOK := orm.CoerceInt64(row["id"])
	if !idOK {
		http.Error(w, "Module not found", http.StatusNotFound)
		return moduleRow{}, 0, false
	}
	return parsed, int(recordID), true
}

func findAppsModule(modules []appsModule, name string) (appsModule, bool) {
	for _, moduleEntry := range modules {
		if moduleEntry.Name == name {
			return moduleEntry, true
		}
	}
	return appsModule{}, false
}
