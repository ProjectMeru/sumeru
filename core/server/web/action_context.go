package web

import (
	"encoding/json"
	"strings"

	"sumeru/core/orm"
)

func parseActionContext(actionData map[string]interface{}) map[string]interface{} {
	raw := strings.TrimSpace(orm.AsString(actionData["context"]))
	if raw == "" {
		return nil
	}
	var ctx map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &ctx); err != nil {
		return nil
	}
	return ctx
}

func actionViewIDFromContext(actionData map[string]interface{}) string {
	if id := strings.TrimSpace(orm.AsString(actionData["view_id"])); id != "" {
		return id
	}
	ctx := parseActionContext(actionData)
	if ctx == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(ctx["view_id"]))
}

func actionSearchViewIDFromContext(actionData map[string]interface{}) string {
	if id := strings.TrimSpace(orm.AsString(actionData["search_view_id"])); id != "" {
		return id
	}
	ctx := parseActionContext(actionData)
	if ctx == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(ctx["search_view_id"]))
}

func actionTarget(actionData map[string]interface{}) string {
	switch strings.ToLower(strings.TrimSpace(orm.AsString(actionData["target"]))) {
	case "new", "dialog":
		return "new"
	case "fullscreen":
		return "fullscreen"
	case "main":
		return "main"
	default:
		return "current"
	}
}

func actionResID(actionData map[string]interface{}) int {
	id, ok := orm.CoerceInt64(actionData["res_id"])
	if !ok || 0 >= id {
		return 0
	}
	return int(id)
}

func actionListLimit(actionData map[string]interface{}) int {
	n, ok := orm.CoerceInt64(actionData["limit"])
	if !ok {
		return workspaceListPageSize
	}
	if n <= 0 {
		return workspaceListPageSize
	}
	if n > maxWorkspaceListRows {
		return maxWorkspaceListRows
	}
	return int(n)
}

// actionDefaultFieldValues extracts default_* keys from action context for new-record forms.
func actionDefaultFieldValues(actionData map[string]interface{}) map[string]interface{} {
	ctx := parseActionContext(actionData)
	if 0 == len(ctx) {
		return nil
	}
	out := make(map[string]interface{})
	for key, value := range ctx {
		if !strings.HasPrefix(key, "default_") {
			continue
		}
		fieldName := strings.TrimPrefix(key, "default_")
		if fieldName == "" {
			continue
		}
		out[fieldName] = value
	}
	if 0 == len(out) {
		return nil
	}
	return out
}
