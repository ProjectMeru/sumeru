package web

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"path/filepath"

	"sumeru/core/engine/render"
	"sumeru/core/server/config"
)

type shellPageOpts struct {
	Route               string
	InnerTemplate       string
	InnerData           interface{}
	MenuIDStr           string
	Page                render.PageData
	ExtraStylesheetURLs []string
}

func renderShellPage(w http.ResponseWriter, r *http.Request, opts shellPageOpts) {
	ctx := r.Context()
	route := resolveShellRoute(opts, r)

	innerHTML, ok := executeInnerTemplate(ctx, w, route, opts)
	if !ok {
		return
	}

	page := finalizeShellPage(ctx, r, opts, innerHTML, route)
	layoutHTML, err := render.RenderPage(ctx, config.AppConfig.TemplatesPath, page)
	if err != nil {
		WebLogEvent(ctx, route, "Failed to render page layout", "render", "failure", err, nil)
		http.Error(w, "Layout render error", http.StatusInternalServerError)
		return
	}
	writeHTML(w, ctx, route, layoutHTML)
}

func resolveShellRoute(opts shellPageOpts, r *http.Request) string {
	if opts.Route != "" {
		return opts.Route
	}
	return r.URL.Path
}

func executeInnerTemplate(ctx context.Context, w http.ResponseWriter, route string, opts shellPageOpts) (template.HTML, bool) {
	templatePaths := []string{
		filepath.Join(config.AppConfig.TemplatesPath, opts.InnerTemplate),
		filepath.Join(config.AppConfig.TemplatesPath, shellPartialsTemplate),
	}
	templateFile, err := template.ParseFiles(templatePaths...)
	if err != nil {
		WebLogEvent(ctx, route, "Failed to parse inner template", "render", "failure", err,
			map[string]interface{}{"template": opts.InnerTemplate})
		http.Error(w, "Template error", http.StatusInternalServerError)
		return "", false
	}

	var innerBuffer bytes.Buffer
	if err := templateFile.Execute(&innerBuffer, opts.InnerData); err != nil {
		WebLogEvent(ctx, route, "Failed to execute inner template", "render", "failure", err, nil)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return "", false
	}
	return template.HTML(innerBuffer.String()), true
}

func finalizeShellPage(ctx context.Context, r *http.Request, opts shellPageOpts, innerHTML template.HTML, route string) render.PageData {
	topMenus, sidebarMenus, activeModuleID, moduleName := render.LoadShellMenus(ctx, opts.MenuIDStr)

	page := opts.Page
	page.Content = innerHTML
	page.TopMenus = topMenus
	page.SidebarMenus = sidebarMenus
	page.ActiveModuleID = activeModuleID
	return applyShellPageDefaults(page, opts, route, r, moduleName, sidebarMenus)
}

func applyShellPageDefaults(page render.PageData, opts shellPageOpts, route string, r *http.Request, moduleName string, sidebarMenus []render.SidebarMenu) render.PageData {
	if page.Title == "" {
		page.Title = defaultPageTitle
	}
	if page.ModuleName == "" {
		page.ModuleName = moduleName
	}
	if page.ActiveMenuID == "" {
		page.ActiveMenuID = opts.MenuIDStr
	}
	if !page.SuppressSidebar && !render.SidebarHasMenus(sidebarMenus) {
		page.SuppressSidebar = true
	}
	if len(page.ViewStylesheetURLs) == 0 {
		page.ViewStylesheetURLs = []string{workspaceStylesheetURL}
	}
	page.ExtraStylesheetURLs = resolveExtraStylesheets(page.ExtraStylesheetURLs, opts.ExtraStylesheetURLs)
	page.ExtraScriptURLs = resolveExtraScripts(page.ExtraScriptURLs, nil)
	if page.CSRFToken == "" {
		page.CSRFToken = CSRFTokenForRequest(r)
	}
	if route == homeRoute {
		page.HomeNavActive = true
	}
	return page
}

func resolveExtraStylesheets(pageStylesheets, optStylesheets []string) []string {
	if len(optStylesheets) > 0 {
		return optStylesheets
	}
	if len(pageStylesheets) > 0 {
		return pageStylesheets
	}
	return render.ExtraStylesheetURLs
}

func resolveExtraScripts(pageScripts, optScripts []string) []string {
	if len(optScripts) > 0 {
		return optScripts
	}
	if len(pageScripts) > 0 {
		return pageScripts
	}
	return render.ExtraScriptURLs
}

func writeHTML(w http.ResponseWriter, ctx context.Context, route, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		WebLogEvent(ctx, route, "Failed to write HTML response", "write", "partial", err, nil)
	}
}
