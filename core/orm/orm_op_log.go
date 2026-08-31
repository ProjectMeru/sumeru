package orm

import (
	"context"
	"fmt"
	"time"

	"sumeru/core/applog"
	"sumeru/core/metrics"
)

func logORMOperation(ctx context.Context, start time.Time, operation, modelName string, err error, ctxFields map[string]interface{}) {
	d := time.Since(start)
	metrics.Inc("sumeru_orm_ops_total")
	metrics.ObserveDuration("sumeru_db_query_duration_seconds", d)

	ctxMap := map[string]interface{}{"resource": modelName}
	for k, v := range ctxFields {
		ctxMap[k] = v
	}

	ev := applog.Event{
		Component: "orm",
		Module:    DeclaringModule(modelName),
		Operation: operation,
		Duration:  d,
		Context:   ctxMap,
		Err:       err,
	}
	if err != nil {
		ev.Message = humanORMMessage(operation, modelName, false)
		ev.Status = "failure"
		applog.Error(ctx, ev)
		return
	}
	ev.Message = humanORMMessage(operation, modelName, true)
	ev.Status = "success"
	if operation == "search" || operation == "search_one" || operation == "search_page" || operation == "find_ui_view" || operation == "find_ui_view_by_name" {
		applog.Debug(ctx, ev)
		return
	}
	applog.Info(ctx, ev)
}

func humanORMMessage(operation, modelName string, success bool) string {
	switch operation {
	case "create":
		if success {
			return "Record created successfully"
		}
		return "Record create failed"
	case "update", "write":
		if success {
			return "Record updated successfully"
		}
		return "Record update failed"
	case "delete", "unlink", "unlink_where":
		if success {
			return "Record deleted successfully"
		}
		return "Record delete failed"
	case "upsert":
		if success {
			return "Record upserted successfully"
		}
		return "Record upsert failed"
	case "search", "search_one", "search_page":
		if success {
			return "Search completed"
		}
		return "Search failed"
	default:
		if success {
			return fmt.Sprintf("%s on %s completed", operation, modelName)
		}
		return fmt.Sprintf("%s on %s failed", operation, modelName)
	}
}

// logORMOperationKV adapts legacy key-value call sites during migration.
func logORMOperationKV(ctx context.Context, start time.Time, operation, modelName string, err error, keysAndValues ...interface{}) {
	ctxMap := map[string]interface{}{}
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		if k, ok := keysAndValues[i].(string); ok {
			ctxMap[k] = keysAndValues[i+1]
		}
	}
	logORMOperation(ctx, start, operation, modelName, err, ctxMap)
}
