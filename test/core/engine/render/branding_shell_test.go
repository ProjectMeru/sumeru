package render_test

import (
	"context"
	"testing"

	"sumeru/core/engine/render"
)

func TestSetShellBranding_andLogoURL(t *testing.T) {
	render.SetShellBranding(render.ShellBranding{
		LogoURL: "/static/logo.png",
		Company: "Acme",
		User:    "Admin",
	})
	if got := render.ShellLogoURL(); got != "/static/logo.png" {
		t.Fatalf("ShellLogoURL: %q", got)
	}
	render.SetShellBranding(render.ShellBranding{})
	if got := render.ShellLogoURL(); got != "" {
		t.Fatalf("cleared logo: %q", got)
	}
}

func TestEnrichShellPageData_noDB(t *testing.T) {
	ctx := context.Background()
	page := render.PageData{Title: "Test"}
	render.EnrichShellPageData(ctx, &page)
	if page.AppName != render.AppDisplayName {
		t.Fatalf("AppName: %q", page.AppName)
	}
	if page.BrandLockupHref == "" || page.HomeNavHref == "" {
		t.Fatalf("nav hrefs: brand=%q home=%q", page.BrandLockupHref, page.HomeNavHref)
	}
	if page.UserProfileHref == "" || page.UserDocsHref == "" {
		t.Fatalf("profile/docs hrefs empty")
	}
	if page.ShellUserInitials == "" && page.ShellUser != "" {
		t.Fatal("expected initials when user set")
	}
}

func TestBuildSwcAddonEntriesJSON(t *testing.T) {
	js := render.BuildSwcAddonEntriesJSON()
	if len(js) == 0 {
		t.Fatal("BuildSwcAddonEntriesJSON empty")
	}
}
