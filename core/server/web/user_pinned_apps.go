package web

import (
	"encoding/json"
	"net/http"

	"sumeru/core/engine/render"
)

type pinnedAppsRequest struct {
	Modules []string `json:"modules"`
}

type pinnedAppsResponse struct {
	Modules []string `json:"modules"`
}

// PinnedAppsSaveHandler persists the user's pinned application modules.
func PinnedAppsSaveHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginJSONPost(w, r) {
		return
	}

	request, ok := parsePinnedAppsRequest(r)
	if !ok {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	savedModules, err := render.SavePinnedAppsForUser(r.Context(), request.Modules)
	if err != nil {
		WebLogEvent(r.Context(), WebLogInput{
			Route: pinnedAppsRoute, Message: "Failed to save pinned apps",
			Operation: "save", Status: "failure", Err: err,
		})
		http.Error(w, "Failed to save pinned apps", http.StatusInternalServerError)
		return
	}

	writeJSON(w, r.Context(), pinnedAppsRoute, pinnedAppsResponse{
		Modules: normalizePinnedModules(savedModules),
	})
}

func parsePinnedAppsRequest(r *http.Request) (pinnedAppsRequest, bool) {
	if acceptsJSONContentType(r.Header.Get("Content-Type")) {
		return parsePinnedAppsJSONBody(r)
	}
	return parsePinnedAppsFormBody(r)
}

func parsePinnedAppsJSONBody(r *http.Request) (pinnedAppsRequest, bool) {
	requestBody, readOK := readBoundedRequestBody(r, maxPinnedAppsBodyBytes)
	if !readOK || int64(len(requestBody)) > maxPinnedAppsBodyBytes {
		return pinnedAppsRequest{}, false
	}

	var request pinnedAppsRequest
	if err := json.Unmarshal(requestBody, &request); err != nil {
		return pinnedAppsRequest{}, false
	}
	return request, true
}

func parsePinnedAppsFormBody(r *http.Request) (pinnedAppsRequest, bool) {
	if err := r.ParseForm(); err != nil {
		return pinnedAppsRequest{}, false
	}

	modules, ok := decodePinnedModulesField(r.FormValue(pinnedAppsModulesField))
	if !ok {
		return pinnedAppsRequest{}, false
	}
	return pinnedAppsRequest{Modules: modules}, true
}

func decodePinnedModulesField(raw string) ([]string, bool) {
	if raw == "" {
		return []string{}, true
	}

	var modules []string
	if err := json.Unmarshal([]byte(raw), &modules); err != nil {
		return nil, false
	}
	return modules, true
}

func normalizePinnedModules(modules []string) []string {
	if modules == nil {
		return []string{}
	}
	return modules
}
