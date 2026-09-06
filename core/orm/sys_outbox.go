package orm

import (
	"context"
	"encoding/json"
	"time"

	"sumeru/core/applog"
	"sumeru/core/errcode"
	"sumeru/core/modelmeta"
)

type SysOutboxEvent struct {
	modelmeta.ModelMeta `sumeru:"model=sys.outbox.event"`

	Name        modelmeta.String `sumeru:"required,index"`
	PayloadJson modelmeta.Text   `sumeru:"column=payload_json"`
	Actor       modelmeta.Integer
	CreatedAt   modelmeta.DateTime `sumeru:"required,column=created_at"`
	PublishedAt modelmeta.DateTime `sumeru:"index,column=published_at"`
}

func outboxValues(name string, actor int, payload map[string]interface{}) map[string]interface{} {
	pj := ""
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			pj = string(b)
		} else {
			applog.WarnCode(context.Background(), errcode.InternalError, "Outbox payload marshal failed", applog.Event{
				Component: "orm",
				Operation: "outbox",
				Status:    "partial",
				Context:   map[string]interface{}{"event_name": name},
				Err:       err,
			})
		}
	}
	return map[string]interface{}{
		"name":         name,
		"payload_json": pj,
		"actor":        actor,
		"created_at":   time.Now().UTC().Format(time.RFC3339),
	}
}

// EnqueueOutboxTx inserts on tx when non-nil.
func EnqueueOutboxTx(ctx context.Context, tx TxWrapper, name string, actor int, payload map[string]interface{}) error {
	if name == "" {
		return nil
	}
	vals := outboxValues(name, actor, payload)
	return insertSideEffectRow(ctx, tx, "sys.outbox.event", vals)
}
