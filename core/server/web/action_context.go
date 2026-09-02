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
	ctx := parseActionContext(actionData)
	if ctx == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(ctx["view_id"]))
}

func actionSearchViewIDFromContext(actionData map[string]interface{}) string {
	ctx := parseActionContext(actionData)
	if ctx == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(ctx["search_view_id"]))
}

// actionDefaultFieldValues extracts default_* keys from action context for new-record forms.
func actionDefaultFieldValues(actionData map[string]interface{}) map[string]interface{} {
	ctx := parseActionContext(actionData)
	if len(ctx) == 0 {
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
	if len(out) == 0 {
		return nil
	}
	return out
}
