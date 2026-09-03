package web_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestAppsRedirectURL(t *testing.T) {
	browse := web.AppsBrowseState{
		Layout:      "grid",
		Filter:      "installed",
		Scope:       "apps",
		SearchQuery: "crm",
	}
	got := web.AppsRedirectURL("installed_sale", browse)
	assertQueryContains(t, got, map[string]string{
		"msg":    "installed_sale",
		"layout": "grid",
		"filter": "installed",
		"scope":  "apps",
		"q":      "crm",
	})

	empty := web.AppsRedirectURL("", web.AppsBrowseState{})
	if empty != web.TestAppsRoute {
		t.Fatalf("got %q want %q", empty, web.TestAppsRoute)
	}
}

func TestAppsRedirectAfterInstallIncludesModule(t *testing.T) {
	browse := web.AppsBrowseState{
		Layout:     "grid",
		Filter:     "all",
		Scope:      "apps",
		ModuleName: "account",
	}
	got := web.AppsRedirectURL("installed_account", browse)
	assertQueryContains(t, got, map[string]string{
		"msg":    "installed_account",
		"module": "account",
		"layout": "grid",
		"scope":  "apps",
	})
}

func TestParseAppsBrowseStateFromForm(t *testing.T) {
	body := "apps_layout=list&apps_filter=installed&apps_scope=technical&apps_q=sale&apps_category=CRM&apps_group_by=category"
	req := httptest.NewRequest("POST", web.TestModuleActionRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	browse := web.ParseAppsBrowseStateFromForm(req)
	if browse.Layout != "list" || browse.Filter != "installed" || browse.Scope != "technical" || browse.SearchQuery != "sale" {
		t.Fatalf("unexpected browse: %+v", browse)
	}
	if browse.Category != "CRM" || browse.GroupBy != web.TestAppsGroupByCategory {
		t.Fatalf("unexpected category/group_by: %+v", browse)
	}
}

func TestAppsDetailRedirectURL(t *testing.T) {
	browse := web.AppsBrowseState{
		Layout:      "grid",
		Filter:      "all",
		Scope:       "all",
		SearchQuery: "",
		ModuleName:  "sale",
	}
	got := web.AppsDetailRedirectURL("saved", browse)
	assertQueryContains(t, got, map[string]string{
		"msg":    "saved",
		"module": "sale",
		"layout": "grid",
	})
}

func TestParseModuleActionForm(t *testing.T) {
	body := "do=install&module=sale&apps_layout=grid&apps_filter=installed&apps_scope=apps&apps_q=crm"
	req := httptest.NewRequest("POST", web.TestModuleActionRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	form := web.ParseModuleActionForm(req)
	if form.Action != web.TestModuleActionInstall || form.ModuleName != "sale" {
		t.Fatalf("unexpected form: %+v", form)
	}
	if form.Browse.Layout != "grid" || form.Browse.Filter != "installed" || form.Browse.Scope != "apps" || form.Browse.SearchQuery != "crm" {
		t.Fatalf("unexpected browse: %+v", form.Browse)
	}
}

func TestRunModuleLifecycleActionUnknown(t *testing.T) {
	_, err := web.RunModuleLifecycleAction(context.Background(), "upgrade", "sale")
	if err == nil || err.Error() != web.TestModuleMsgUnknownAction {
		t.Fatalf("got err %v want %q", err, web.TestModuleMsgUnknownAction)
	}
}
