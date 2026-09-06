package orm

import (
	"context"
	"fmt"

	"sumeru/core/applog"
	"sumeru/core/errcode"
	"sumeru/core/event"
)

func shouldEmitSideEffects(ctx context.Context, modelName string) bool {
	return !SecurityBypass(ctx) && !skipAuditModel(modelName)
}

type sideEffectRow struct {
	Action    string
	EventName string
	ResID     int64
	Before    map[string]interface{}
	After     map[string]interface{}
	Detail    string
}

func emitSideEffectsOnTx(ctx context.Context, tx TxWrapper, modelName string, uid int, rows []sideEffectRow) {
	if !shouldEmitSideEffects(ctx, modelName) {
		return
	}
	for _, row := range rows {
		AppendAuditTx(ctx, tx, row.Action, modelName, row.ResID, row.Before, row.After, row.Detail)
		if err := EnqueueOutboxTx(ctx, tx, row.EventName, uid, map[string]interface{}{
			"model": modelName,
			"id":    int(row.ResID),
		}); err != nil {
			logSideEffectWarn(ctx, "outbox_enqueue", modelName, err, "event", row.EventName)
		}
	}
}

func publishRecordEvents(ctx context.Context, eventName string, uid int, modelName string, ids []int) {
	if !shouldEmitSideEffects(ctx, modelName) {
		return
	}
	for _, id := range ids {
		if errs := event.Publish(ctx, event.Event{
			Name:    eventName,
			Actor:   uid,
			Payload: map[string]interface{}{"model": modelName, "id": id},
		}); len(errs) > 0 {
			logSideEffectWarn(ctx, "publish", modelName, fmt.Errorf("%v", errs), "resource_id", id)
		}
	}
}

func logSideEffectWarn(ctx context.Context, operation, modelName string, err error, extra ...interface{}) {
	ctxMap := map[string]interface{}{"resource": modelName}
	for i := 0; i+1 < len(extra); i += 2 {
		if k, ok := extra[i].(string); ok {
			ctxMap[k] = extra[i+1]
		}
	}
	applog.WarnCode(ctx, errcode.InternalError, operation+" side effect failed for "+modelName, applog.Event{
		Component: "orm",
		Module:    DeclaringModule(modelName),
		Operation: operation,
		Status:    "partial",
		Context:   ctxMap,
		Err:       err,
	})
}
