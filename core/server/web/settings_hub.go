package web

import (
	"context"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

type settingsHubLink struct {
	Name string
	Href string
}

type settingsHubSection struct {
	Title      string
	FilterText string
	Links      []settingsHubLink
}

type settingsHubAppTile = render.AppTile

type settingsHubData struct {
	Sections          []settingsHubSection
	AppTiles          []settingsHubAppTile
	CompaniesMenuHref string
}

// SettingsHubHandler renders the Settings overview at /web/settings.
func SettingsHubHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSettingsUser(w, r) {
		return
	}

	ctx := r.Context()
	menuIDStr, ok := resolveSettingsRootMenuID(w, r, ctx)
	if !ok {
		return
	}

	sections := loadSettingsHubSections(ctx, menuIDStr)
	appTiles, err := loadInstalledAppTiles(ctx, false)
	if err != nil {
		WebLogEvent(ctx, WebLogInput{
			Route: settingsRoute, Message: "Failed to list modules for settings hub",
			Operation: "load", Status: "failure", Err: err,
		})
		http.Error(w, "Failed to list modules", http.StatusInternalServerError)
		return
	}

	renderSettingsHubPage(w, r, settingsHubData{
		Sections:          sections,
		AppTiles:          appTiles,
		CompaniesMenuHref: menuHrefFromXMLID(ctx, settingsCompaniesMenuXMLID),
	}, menuIDStr)
}

func requireSettingsUser(w http.ResponseWriter, r *http.Request) bool {
	ctx := r.Context()
	if orm.UserHasGroupXML(ctx, orm.SecurityUID(ctx), groupUserXML) {
		return true
	}
	http.Redirect(w, r, homeRoute, http.StatusFound)
	return false
}

func resolveSettingsRootMenuID(w http.ResponseWriter, r *http.Request, ctx context.Context) (menuIDStr string, ok bool) {
	_, menuIDStr = resolveMenuID(ctx, settingsHubMenuXMLID)
	if menuIDStr != "" {
		return menuIDStr, true
	}
	http.Redirect(w, r, appsRoute, http.StatusFound)
	return "", false
}

func loadSettingsHubSections(ctx context.Context, menuIDStr string) []settingsHubSection {
	_, sidebarMenus, _, _ := render.LoadShellMenus(ctx, menuIDStr)
	sections := make([]settingsHubSection, 0, len(sidebarMenus))
	for _, sidebarSection := range sidebarMenus {
		if hubSection, include := settingsHubSectionFromSidebar(sidebarSection); include {
			sections = append(sections, hubSection)
		}
	}
	return sections
}

func settingsHubSectionFromSidebar(sidebarSection render.SidebarMenu) (settingsHubSection, bool) {
	sectionTitle := strings.TrimSpace(sidebarSection.Name)
	filterTerms := []string{strings.ToLower(sectionTitle)}
	links := make([]settingsHubLink, 0, len(sidebarSection.SubMenus))

	for _, subMenu := range sidebarSection.SubMenus {
		linkName := strings.TrimSpace(subMenu.Name)
		linkHref := strings.TrimSpace(subMenu.Action)
		if linkName == "" || linkHref == "" {
			continue
		}
		links = append(links, settingsHubLink{Name: linkName, Href: linkHref})
		filterTerms = append(filterTerms, strings.ToLower(linkName))
	}

	if len(links) == 0 {
		return settingsHubSection{}, false
	}

	return settingsHubSection{
		Title:      sectionTitle,
		FilterText: strings.Join(filterTerms, " "),
		Links:      links,
	}, true
}

func renderSettingsHubPage(w http.ResponseWriter, r *http.Request, pageData settingsHubData, menuIDStr string) {
	ctx := r.Context()
	renderShellPage(w, r, shellPageOpts{
		Route:         settingsRoute,
		InnerTemplate: settingsHubInnerTemplate,
		InnerData:     pageData,
		MenuIDStr:     menuIDStr,
		Page:          buildSettingsHubPageData(ctx, menuIDStr),
	})
}

func buildSettingsHubPageData(ctx context.Context, menuIDStr string) render.PageData {
	return render.PageData{
		Title:                settingsHubPageTitle,
		SettingsNavActive:    true,
		ActiveMenuID:         menuIDStr,
		SuppressActivityDock: true,
		BreadcrumbItems:      render.BuildSettingsHubBreadcrumbs(ctx),
		ViewStylesheetURLs:   []string{settingsHubStylesheetURL},
		ExtraBodyClasses:     settingsHubBodyClass,
	}
}
