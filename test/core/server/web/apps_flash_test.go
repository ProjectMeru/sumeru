package web_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/server/web"
)

func TestAppsFlashFromMessageInstalled(t *testing.T) {
	names := map[string]string{"account": "Accounting"}
	flash, ok := web.AppsFlashFromMessage("installed_account", names)
	if !ok {
		t.Fatal("expected flash")
	}
	if flash.Kind != "success" || flash.Title != "App installed" {
		t.Fatalf("flash = %+v", flash)
	}
	if flash.Body != "Accounting is ready to use." {
		t.Fatalf("body = %q", flash.Body)
	}
	if !flash.ToastOnly {
		t.Fatal("expected ToastOnly for install success")
	}
}

func TestAppsFlashFromMessageUninstalled(t *testing.T) {
	names := map[string]string{"sale": "Sales"}
	flash, ok := web.AppsFlashFromMessage("uninstalled_sale", names)
	if !ok || !flash.ToastOnly || flash.Title != "App removed" {
		t.Fatalf("flash = %+v ok=%v", flash, ok)
	}
}

func TestSplitAppsPageFlashesErrorInline(t *testing.T) {
	inline, toast := web.SplitAppsPageFlashes("error:depends on missing", nil)
	if len(toast) != 0 {
		t.Fatalf("unexpected toast: %+v", toast)
	}
	if len(inline) != 1 || inline[0].Kind != "error" {
		t.Fatalf("inline = %+v", inline)
	}
}

func TestSplitAppsPageFlashesInstallToast(t *testing.T) {
	inline, toast := web.SplitAppsPageFlashes("installed_crm", map[string]string{"crm": "CRM"})
	if len(inline) != 0 {
		t.Fatalf("unexpected inline: %+v", inline)
	}
	if len(toast) != 1 || !toast[0].ToastOnly {
		t.Fatalf("toast = %+v", toast)
	}
}

func TestFormatAppsActionError(t *testing.T) {
	if got := web.FormatAppsActionError("depends on x"); got != "error:depends on x" {
		t.Fatalf("got %q", got)
	}
	if got := web.FormatAppsActionError("error:already"); got != "error:already" {
		t.Fatalf("got %q", got)
	}
}

func TestEnrichAppsModulesDetailURL(t *testing.T) {
	modules := []web.AppsModule{{Name: "account", DisplayName: "Accounting"}}
	browse := web.AppsBrowseState{Layout: "grid", Filter: "all", Scope: "apps"}
	enriched := web.EnrichAppsModules(context.Background(), modules, browse)
	if len(enriched) != 1 {
		t.Fatalf("len = %d", len(enriched))
	}
	if enriched[0].DetailURL == "" {
		t.Fatal("expected DetailURL")
	}
	if !strings.Contains(enriched[0].DetailURL, "module=account") {
		t.Fatalf("DetailURL = %q", enriched[0].DetailURL)
	}
}

func TestModuleStatusLabel(t *testing.T) {
	if got := web.ModuleStatusLabel(web.ModuleRow{State: "installed", Active: true}); got != "Installed" {
		t.Fatalf("got %q", got)
	}
	if got := web.ModuleStatusLabel(web.ModuleRow{State: "installed", Active: false}); got != "Installed (inactive)" {
		t.Fatalf("got %q", got)
	}
	if got := web.ModuleStatusLabel(web.ModuleRow{State: "uninstalled"}); got != "Not installed" {
		t.Fatalf("got %q", got)
	}
}
