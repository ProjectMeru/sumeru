package automation

import (
	"context"
	"encoding/json"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/orm"
)

// executeServerAction runs a matching sys.server.action row for an event.
// Code conventions (declarative, no script engine):
//   - publish:<event_name>  — publish a follow-up event (payload copied from trigger)
//   - write:<json_object>   — ORM write on payload model/id
func executeServerAction(ctx context.Context, row map[string]interface{}, ev event.Event) error {
	actionModel := strings.TrimSpace(orm.AsString(row["model"]))
	if actionModel != "" {
		evModel := strings.TrimSpace(orm.AsString(ev.Payload["model"]))
		if evModel != "" && actionModel != evModel {
			return nil
		}
	}
	if domRaw := strings.TrimSpace(orm.AsString(row["trigger_domain"])); domRaw != "" {
		modelName := strings.TrimSpace(orm.AsString(ev.Payload["model"]))
		resID, ok := orm.CoerceInt64(ev.Payload["id"])
		if modelName == "" || !ok || resID <= 0 {
			return nil
		}
		var domain [][]interface{}
		if err := json.Unmarshal([]byte(domRaw), &domain); err != nil {
			return nil
		}
		matches, err := orm.SearchCount(ctx, modelName, append(domain, []interface{}{"id", "=", resID}))
		if err != nil || matches == 0 {
			return nil
		}
	}

	code := strings.TrimSpace(orm.AsString(row["code"]))
	if code == "" {
		return nil
	}

	switch {
	case strings.HasPrefix(code, "publish:"):
		name := strings.TrimSpace(strings.TrimPrefix(code, "publish:"))
		if name == "" {
			return nil
		}
		payload := ev.Payload
		if payload == nil {
			payload = map[string]interface{}{}
		}
		errs := event.Publish(ctx, event.Event{Name: name, Actor: ev.Actor, Payload: payload})
		if len(errs) > 0 {
			return errs[0]
		}
		return nil

	case strings.HasPrefix(code, "write:"):
		raw := strings.TrimSpace(strings.TrimPrefix(code, "write:"))
		if raw == "" {
			return nil
		}
		var vals map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &vals); err != nil {
			applog.Warn(ctx, applog.Event{
				Message:   "server action write JSON invalid",
				Component: "automation",
				Operation: "server_action",
				Status:    "failed",
				Context:   map[string]interface{}{"action": row["name"]},
				Err:       err,
			})
			return nil
		}
		modelName := strings.TrimSpace(orm.AsString(ev.Payload["model"]))
		resID, ok := orm.CoerceInt64(ev.Payload["id"])
		if modelName == "" || !ok || resID <= 0 {
			return nil
		}
		bypass := orm.ContextWithBypass(ctx, true)
		return orm.UpdateRecordByID(bypass, modelName, int(resID), vals)

	case strings.HasPrefix(code, "webhook:"):
		url := strings.TrimSpace(strings.TrimPrefix(code, "webhook:"))
		if url == "" {
			return nil
		}
		return dispatchWebhook(ctx, url, ev)

	default:
		applog.DebugMsg(ctx, "module", "automation",
			"server action code not executed (unknown prefix)",
			map[string]interface{}{"action": row["name"], "code": code})
	}
	return nil
}
