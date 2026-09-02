package web_test

import (
	"net/url"
	"sumeru/core/server/web"
	"testing"
)

func TestGroupAppsModulesByCategory(t *testing.T) {
	modules := []web.AppsModule{
		{Name: "sale", DisplayName: "Sales", Category: "Sales"},
		{Name: "crm", DisplayName: "CRM", Category: "CRM"},
		{Name: "account", DisplayName: "Accounting", Category: "Accounting"},
	}
	groups := web.GroupAppsModules(modules, web.TestAppsGroupByCategory)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Label != "Accounting" || len(groups[0].Modules) != 1 {
		t.Fatalf("first group: %+v", groups[0])
	}
	if groups[2].Label != "Sales" {
		t.Fatalf("last group label: %q", groups[2].Label)
	}
}

func TestFilterAppsModulesByCategory(t *testing.T) {
	modules := []web.AppsModule{
		{Name: "sale", DisplayName: "Sales", Category: "Sales", Application: true},
		{Name: "crm", DisplayName: "CRM", Category: "CRM", Application: true},
	}
	browse := web.AppsBrowseState{Filter: "all", Scope: "apps", Category: "CRM"}
	appModules, _ := web.FilterAppsModulesByBrowse(modules, browse)
	if len(appModules) != 1 || appModules[0].Name != "crm" {
		t.Fatalf("appModules: %+v", appModules)
	}
}

func TestAppsLinkPreservesCategoryAndGroupBy(t *testing.T) {
	got := web.AppsLinkFromBrowse(web.AppsBrowseState{
		Layout:      "list",
		Filter:      "installed",
		Scope:       "apps",
		SearchQuery: "crm",
		Category:    "CRM",
		GroupBy:     web.TestAppsGroupByCategory,
	})
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	for key, value := range map[string]string{
		"layout":   "list",
		"filter":   "installed",
		"scope":    "apps",
		"q":        "crm",
		"category": "CRM",
		"group_by": "category",
	} {
		if query.Get(key) != value {
			t.Fatalf("got %s=%q want %q", key, query.Get(key), value)
		}
	}
}

func TestModuleSummaryPrefersDescriptionFirstLine(t *testing.T) {
	got := web.ModuleSummaryFromDescription("Short line\nLong body")
	if got != "Short line" {
		t.Fatalf("got %q", got)
	}
}

func TestModuleHasLongDescription(t *testing.T) {
	if web.ModuleHasLongDescription("Same", "Same") {
		t.Fatal("identical summary and description should not be long")
	}
	long := "This is a much longer description that goes on and on with extra detail beyond the summary line."
	if !web.ModuleHasLongDescription("Brief.", long) {
		t.Fatal("expected long description")
	}
}
