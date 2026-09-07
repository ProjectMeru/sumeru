package orm

import (
	"context"
	"fmt"
	"time"
)

// UpdateRecordByID sets columns from prepared values for a single id.
func UpdateRecordByID(ctx context.Context, modelName string, id int, values map[string]interface{}) (err error) {
	start := time.Now()
	defer func() {
		logORMOperation(ctx, start, "write", modelName, err, map[string]interface{}{"resource_id": id})
	}()
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	_, err = Update(ctx, modelName, [][]interface{}{{"id", "=", id}}, values)
	if err == nil && modelName == "core.user" {
		if v, ok := values["active"]; ok && !AsBool(v) {
			DestroySessionsForUser(ctx, id)
		}
	}
	return err
}

// Update updates rows matching domain with values inside a mandatory transaction.
func Update(ctx context.Context, modelName string, domain [][]interface{}, values map[string]interface{}) (n int64, err error) {
	start := time.Now()
	defer func() {
		logORMOperation(ctx, start, "update", modelName, err, map[string]interface{}{"rows": n})
	}()
	res, err := executeUpdateMutation(ctx, modelName, domain, values)
	return res.RowsAffected, err
}

func mergeRecordMap(base map[string]interface{}, patch map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range base {
		out[k] = v
	}
	for k, v := range patch {
		out[k] = v
	}
	return out
}
