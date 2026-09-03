package web_test

import (
	"sumeru/core/server/web"
	"testing"
)

func TestAppsModuleFromParsed(t *testing.T) {
	installed := web.AppsModuleFromParsed(web.ModuleRow{
		Name: "sale", State: "installed", Active: true, Application: true,
	})
	if installed.CanInstall || !installed.CanUninstall || !installed.CanDeactivate || installed.CanActivate {
		t.Fatalf("installed active app: %+v", installed)
	}

	uninstalled := web.AppsModuleFromParsed(web.ModuleRow{Name: "crm", State: "uninstalled", Application: true})
	if !uninstalled.CanInstall || uninstalled.CanUninstall {
		t.Fatalf("uninstalled app: %+v", uninstalled)
	}

	base := web.AppsModuleFromParsed(web.ModuleRow{Name: "base", State: "installed", Active: true, Application: true})
	if base.CanUninstall || base.CanDeactivate || base.CanActivate {
		t.Fatalf("base module must not be lifecycle-managed: %+v", base)
	}
}

func TestFilterAppsModulesByBrowse(t *testing.T) {
	modules := []web.AppsModule{
		{Name: "sale", DisplayName: "Sales", State: "installed", Application: true},
		{Name: "web", DisplayName: "Web", State: "installed", Application: false},
	}
	browse := web.AppsBrowseState{Filter: "all", Scope: "apps", SearchQuery: "sale"}
	appModules, techModules := web.FilterAppsModulesByBrowse(modules, browse)
	if len(appModules) != 1 || appModules[0].Name != "sale" {
		t.Fatalf("appModules: %+v", appModules)
	}
	if len(techModules) != 0 {
		t.Fatalf("techModules should be cleared for apps scope: %+v", techModules)
	}
}
