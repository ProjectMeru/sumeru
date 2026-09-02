package web

import (
	"encoding/json"
	"net/http"
	"strings"
)

const swcWorkspaceRoute = "/web/swc/workspace"

func registerSwcRoutes() {
	registerSession(http.MethodGet, swcWorkspaceRoute, SwcWorkspaceHandler)
	registerSwcSavedSearchRoutes()
	registerSwcBusRoute()
	registerSwcChatterRoute()
}

// SwcWorkspaceHandler GET /web/swc/workspace — JSON workspace payload for SWC.
func SwcWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
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
	} else {
		http.Error(w, "action required", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), httpStatusFromWorkspaceError(err))
		return
	}
	req := parseWorkspaceRequest(r, actionID)
	req.menuID = CanonicalMenuID(ctx, req.menuID, actionID)
	viewRecord, err := buildViewRecordData(ctx, w, r, req, resolved, actionData)
	if err != nil {
		respondWorkspaceLoadError(w, ctx, err)
		return
	}
	payload := buildSwcWorkspacePayload(ctx, resolved, req, viewRecord, actionData)
	writeJSONResponse(w, payload)
}

func writeJSONResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
