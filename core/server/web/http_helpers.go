package web

import (
	"net/http"
	"strings"

	"sumeru/core/engine/render"
)

func splitViewModes(viewMode string) []string {
	return splitCommaSeparatedValues(viewMode)
}

func splitCommaSeparatedValues(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func firstGroupByField(raw string) string {
	fields := splitCommaSeparatedValues(raw)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func normalizeViewMode(viewMode string) string {
	return strings.ToLower(strings.TrimSpace(viewMode))
}

// formBaseQueryValues builds a stable /web query string (no leading "?") for form Edit/Cancel/Save redirects.
func formBaseQueryValues(actionID int, menuID, viewType, recordID string) string {
	return render.WorkspaceQueryString(render.WorkspaceQuery{
		ActionID: actionID,
		MenuID:   menuID,
		ViewType: viewType,
		RecordID: recordID,
	})
}

func formOrQueryValue(r *http.Request, field string) string {
	if value := strings.TrimSpace(r.FormValue(field)); value != "" {
		return value
	}
	return strings.TrimSpace(r.URL.Query().Get(field))
}

// workspaceListURL builds a /web list redirect after delete (action + menu_id only).
func workspaceListURL(actionID, menuID string) string {
	return render.WorkspaceURL(render.WorkspaceQuery{Action: actionID, MenuID: menuID})
}
