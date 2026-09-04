package orm

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ActionModels lists navigable action tables referenced from sys.menu.action_id.
var ActionModels = []string{"sys.action.window", "sys.action.url"}

// ResolveActionRecord returns the action model name and core row id for a menu /web link.
// actionQuery may be a numeric id or an XML id (module.name); actionID is the parsed numeric id when known.
func ResolveActionRecord(ctx context.Context, actionID int, actionQuery string) (modelName string, coreID int, err error) {
	actionQuery = strings.TrimSpace(actionQuery)
	if actionQuery != "" && strings.Contains(actionQuery, ".") {
		id, model, err := ResolveXmlId(ctx, actionQuery)
		if err != nil {
			return "", 0, err
		}
		if id <= 0 {
			return "", 0, fmt.Errorf("action %q not found", actionQuery)
		}
		return model, id, nil
	}
	if actionID > 0 {
		return actionModelForCoreID(ctx, actionID)
	}
	if actionQuery != "" {
		if id, err := strconv.Atoi(actionQuery); err == nil && id > 0 {
			return actionModelForCoreID(ctx, id)
		}
	}
	return "", 0, fmt.Errorf("action not found")
}

func actionModelForCoreID(ctx context.Context, coreID int) (string, int, error) {
	if coreID <= 0 {
		return "", 0, fmt.Errorf("action not found")
	}
	rows, err := Search(ctx, "sys.model.data", [][]interface{}{
		{"core_id", "=", coreID},
		{"model", "in", actionModelsSlice()},
	})
	if err != nil {
		return "", 0, err
	}
	if len(rows) == 0 {
		return "", 0, fmt.Errorf("action %d not found", coreID)
	}
	modelName := strings.TrimSpace(AsString(rows[0]["model"]))
	if modelName == "" {
		return "", 0, fmt.Errorf("action %d not found", coreID)
	}
	return modelName, coreID, nil
}

func actionModelsSlice() []interface{} {
	out := make([]interface{}, len(ActionModels))
	for i, m := range ActionModels {
		out[i] = m
	}
	return out
}
