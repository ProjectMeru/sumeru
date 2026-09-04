package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/swcmeta"
)

const (
	swcSavedSearchesRoute = "/web/swc/saved-searches"
)

func registerSwcSavedSearchRoutes() {
	registerSession(http.MethodGet, swcSavedSearchesRoute, SwcSavedSearchesListHandler)
	registerSession(http.MethodPost, swcSavedSearchesRoute, SwcSavedSearchSaveHandler)
	registerSession(http.MethodDelete, swcSavedSearchesRoute, SwcSavedSearchDeleteHandler)
}

func SwcSavedSearchesListHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()
	actionID, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get(workspaceActionParam)))
	model := strings.TrimSpace(r.URL.Query().Get(workspaceModelParam))
	writeJSONResponse(w, swcmeta.LoadSavedSearches(ctx, actionID, model))
}

func SwcSavedSearchSaveHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginJSONPost(w, r) {
		return
	}
	var body struct {
		ActionID  int    `json:"actionId"`
		Model     string `json:"model"`
		Name      string `json:"name"`
		Search    string `json:"search"`
		Filter    string `json:"filter"`
		Domain    string `json:"domain"`
		GroupBy   string `json:"groupBy"`
		IsDefault bool   `json:"isDefault"`
		IsShared  bool   `json:"isShared"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	id, err := swcmeta.SaveSavedSearch(r.Context(), swcmeta.SavedSearchInput{
		ActionID:  body.ActionID,
		Model:     body.Model,
		Name:      body.Name,
		Search:    body.Search,
		Filter:    body.Filter,
		Domain:    body.Domain,
		GroupBy:   body.GroupBy,
		IsDefault: body.IsDefault,
		IsShared:  body.IsShared,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSONResponse(w, map[string]int{"id": id})
}

func SwcSavedSearchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, methodNotAllowedMessage, http.StatusMethodNotAllowed)
		return
	}
	if !validateSessionCSRF(w, r) {
		return
	}
	id, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := swcmeta.DeleteSavedSearch(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
