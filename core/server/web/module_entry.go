package web

import (
	"context"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

const moduleWebPrefix = "/web/"

func registerModuleWebRoutes(mux *http.ServeMux) {
	mux.HandleFunc(moduleWebPrefix, moduleWebPrefixHandler)
}

func moduleWebPrefixHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !requireLogin(w, r) {
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, moduleWebPrefix)
	if rest == "" || strings.Contains(rest, "/") {
		http.NotFound(w, r)
		return
	}
	moduleName := strings.TrimSpace(rest)
	if !render.ValidModuleWebSegment(moduleName) {
		http.NotFound(w, r)
		return
	}
	moduleEntryHandler(w, r, moduleName)
}

func moduleEntryHandler(w http.ResponseWriter, r *http.Request, moduleName string) {
	ctx := r.Context()
	if !isOpenApplicationModule(ctx, moduleName) {
		http.NotFound(w, r)
		return
	}
	menuID := render.RootMenuIDForModule(ctx, moduleName)
	if 0 >= menuID {
		http.Redirect(w, r, homeRoute, http.StatusFound)
		return
	}
	uid := orm.SecurityUID(ctx)
	actionID, leafMenuID := orm.MenuEntryActionForUser(ctx, uid, menuID)
	if actionID == 0 {
		http.Redirect(w, r, homeRoute, http.StatusFound)
		return
	}
	loc := render.WorkspaceURLFromMenuAction(ctx, actionID, menuID, leafMenuID)
	if loc == render.HomeWebURL() {
		http.Redirect(w, r, homeRoute, http.StatusFound)
		return
	}
	debugEnabled := shellDebugEnabled(r)
	loc = render.WorkspaceURLWithDebug(loc, debugEnabled)
	http.Redirect(w, r, loc, http.StatusFound)
}

func isOpenApplicationModule(ctx context.Context, moduleName string) bool {
	row, err := orm.SearchOne(ctx, "sys.module", map[string]interface{}{"name": moduleName})
	if err != nil {
		return false
	}
	parsed, ok := parseModuleRow(row)
	if !ok {
		return false
	}
	return parsed.State == "installed" && parsed.Application && parsed.Active
}

func shellDebugEnabled(r *http.Request) bool {
	enabled, _ := render.ParseDebugQuery(r.URL.Query().Get("debug"))
	return enabled
}

func redirectDebugModuleEntry(w http.ResponseWriter, r *http.Request, moduleName string) {
	moduleName = strings.TrimSpace(moduleName)
	if href := render.ModuleWebURL(moduleName); href != "" {
		if shellDebugEnabled(r) {
			href = render.ModuleWebURLWithDebugAlias(moduleName)
		}
		http.Redirect(w, r, href, http.StatusFound)
		return
	}
	http.Redirect(w, r, homeRoute, http.StatusFound)
}

func handleWebDebugQueryRedirect(w http.ResponseWriter, r *http.Request) bool {
	enabled, mod := render.ParseDebugQuery(r.URL.Query().Get("debug"))
	if !enabled {
		return false
	}
	if mod != "" {
		redirectDebugModuleEntry(w, r, mod)
		return true
	}
	return false
}
