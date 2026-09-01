package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// WebHandler renders sys.action.window targets using sys.view definitions.
func WebHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}

	ctx := r.Context()
	actionQuery, menuQuery := workspaceQueryParams(r)

	if redirectIfMenuAccessDenied(w, r, menuQuery) {
		return
	}

	actionID := ResolveWindowActionID(ctx, actionQuery, menuQuery)
	modelQuery := strings.TrimSpace(r.URL.Query().Get(workspaceModelParam))

	var actionData map[string]interface{}
	var resolved *resolvedWorkspaceView
	var err error
	if actionID != 0 {
		actionData, err = loadWindowAction(ctx, actionID)
		if err != nil {
			respondActionNotFound(w, actionID)
			return
		}
		resolved, err = resolveWorkspaceView(ctx, r, actionData)
	} else if modelQuery != "" {
		actionData = map[string]interface{}{}
		resolved, err = resolveWorkspaceByModel(ctx, r, modelQuery)
	} else if redirectIfNoWindowAction(w, r, actionQuery, menuQuery, actionID) {
		return
	} else {
		return
	}
	if err != nil {
		http.Error(w, err.Error(), httpStatusFromWorkspaceError(err))
		return
	}

	req := parseWorkspaceRequest(r, actionID)
	viewRecord, err := buildViewRecordData(ctx, w, r, req, resolved, actionData)
	if err != nil {
		respondWorkspaceLoadError(w, ctx, err)
		return
	}

	html := render.RenderSWCWorkspace(ctx, render.SWCPageInput{
		View:         resolved.view,
		ActiveMenuID: req.menuID,
		TemplatesDir: config.AppConfig.TemplatesPath,
		RecData:      viewRecord,
		SelectedMode: resolved.selectedMode,
	})
	logWorkspaceViewOpened(ctx, r.URL.Path, req, actionID, resolved)
	writeHTML(w, ctx, r.URL.Path, html)
}

func workspaceQueryParams(r *http.Request) (actionQuery, menuQuery string) {
	query := r.URL.Query()
	return query.Get(workspaceActionParam), query.Get(workspaceMenuIDParam)
}

func redirectIfMenuAccessDenied(w http.ResponseWriter, r *http.Request, menuQuery string) bool {
	menuID := strings.TrimSpace(menuQuery)
	if menuID == "" || menuAccessAllowed(r.Context(), menuID) {
		return false
	}

	WebLogf(r.Context(), workspaceRoute, "menu_id=%s denied by access_groups", menuID)
	http.Redirect(w, r, homeRoute, http.StatusFound)
	return true
}

func redirectIfNoWindowAction(w http.ResponseWriter, r *http.Request, actionQuery, menuQuery string, actionID int) bool {
	redirectURL, shouldRedirect := resolveWorkspaceRedirect(r.Context(), menuQuery, actionID)
	if !shouldRedirect {
		return false
	}

	if actionID == 0 && redirectURL == appsRoute {
		WebLogf(r.Context(), workspaceRoute, "no action for query action=%q menu_id=%q; redirecting to apps", actionQuery, menuQuery)
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
	return true
}

func loadWindowAction(ctx context.Context, actionID int) (map[string]interface{}, error) {
	return orm.SearchOne(ctx, sysActionWindowModel, map[string]interface{}{"id": actionID})
}

func respondActionNotFound(w http.ResponseWriter, actionID int) {
	http.Error(w, fmt.Sprintf("Action %d not found", actionID), http.StatusNotFound)
}

func respondWorkspaceLoadError(w http.ResponseWriter, ctx context.Context, err error) {
	WebLogf(ctx, workspaceRoute, "load view data: %v", err)
	http.Error(w, err.Error(), httpStatusFromWorkspaceError(err))
}

func httpStatusFromWorkspaceError(err error) int {
	message := err.Error()
	switch {
	case strings.Contains(message, workspaceErrInvalidID):
		return http.StatusBadRequest
	case strings.Contains(message, workspaceErrNoView), strings.Contains(message, workspaceErrNotFound):
		return http.StatusNotFound
	case strings.Contains(message, "access denied"):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func logWorkspaceViewOpened(ctx context.Context, route string, req workspaceRequest, actionID int, resolved *resolvedWorkspaceView) {
	recordID, _ := parsePositiveRecordID(req.recordID)
	WebLogNavigation(ctx, route, workspaceViewOpenOp, "Workspace view opened", map[string]interface{}{
		"menu_id":   req.menuID,
		"action_id": actionID,
		"model":     resolved.targetModel,
		"view_type": resolved.selectedMode,
		"record_id": recordID,
		"edit":      req.formEdit,
	})
}
