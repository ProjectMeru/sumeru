package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/orm"
)

func actionApplies(ctx context.Context, row map[string]interface{}, ev event.Event) bool {
	actionModel := strings.TrimSpace(orm.AsString(row["model"]))
	evModel := strings.TrimSpace(orm.AsString(ev.Payload["model"]))
	if ev.Name == "cron.tick" {
		if actionModel != "" || strings.TrimSpace(orm.AsString(row["trigger_domain"])) != "" {
			return false
		}
	}
	if actionModel != "" && evModel != "" && actionModel != evModel {
		return false
	}
	domRaw := strings.TrimSpace(orm.AsString(row["trigger_domain"]))
	if ev.Name == orm.EventRecordDeleted {
		if domRaw == "" {
			return true
		}
		actionID, _ := orm.CoerceInt64(row["id"])
		resID, ok := orm.CoerceInt64(ev.Payload["id"])
		return ok && deleteActionMatched(evModel, int(resID), actionID)
	}
	if domRaw == "" {
		return true
	}
	resID, ok := orm.CoerceInt64(ev.Payload["id"])
	if evModel == "" || !ok || 0 >= resID {
		return false
	}
	var domain [][]interface{}
	if err := json.Unmarshal([]byte(domRaw), &domain); err != nil {
		applog.Warn(ctx, applog.Event{
			Message:   "server action domain JSON invalid",
			Component: "automation",
			Operation: "server_action",
			Status:    "failed",
			Context:   map[string]interface{}{"action": row["name"]},
			Err:       err,
		})
		return false
	}
	matches, err := orm.SearchCount(ctx, evModel, append(domain, []interface{}{"id", "=", resID}))
	return err == nil && matches > 0
}

// executeServerAction runs a matching sys.server.action row for an event.
// Code conventions (declarative, no script engine):
//   - publish:<event_name>  — publish a follow-up event (payload copied from trigger)
//   - write:<json_object>   — ORM write on payload model/id
//   - webhook:<https url>
//   - activity:<xmlid>|<summary>
func executeServerAction(ctx context.Context, row map[string]interface{}, ev event.Event) error {
	if !actionApplies(ctx, row, ev) {
		return nil
	}
	code := strings.TrimSpace(orm.AsString(row["code"]))
	if code == "" {
		return nil
	}

	switch {
	case strings.HasPrefix(code, "publish:"):
		if depthOf(ctx) >= automationDepthCap {
			return fmt.Errorf("automation depth exceeded")
		}
		name := strings.TrimSpace(strings.TrimPrefix(code, "publish:"))
		if name == "" {
			return nil
		}
		payload := ev.Payload
		if payload == nil {
			payload = map[string]interface{}{}
		}
		errs := event.Publish(withDepth(ctx, depthOf(ctx)+1), event.Event{Name: name, Actor: ev.Actor, Payload: payload})
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
		if modelName == "" || !ok || 0 >= resID {
			return nil
		}
		if err := rejectUnsafeWrite(modelName, vals); err != nil {
			return err
		}
		return orm.UpdateRecordByID(ctx, modelName, int(resID), vals)

	case strings.HasPrefix(code, "webhook:"):
		url := strings.TrimSpace(strings.TrimPrefix(code, "webhook:"))
		if url == "" {
			return nil
		}
		return dispatchWebhook(ctx, url, ev)

	case strings.HasPrefix(code, "activity:"):
		return createActivity(ctx, strings.TrimPrefix(code, "activity:"), ev)

	default:
		applog.DebugMsg(ctx, "module", "automation",
			"server action code not executed (unknown prefix)",
			map[string]interface{}{"action": row["name"], "code": code})
	}
	return nil
}

func createActivity(ctx context.Context, spec string, ev event.Event) error {
	spec = strings.TrimSpace(spec)
	xmlid, summary, _ := strings.Cut(spec, "|")
	xmlid = strings.TrimSpace(xmlid)
	summary = strings.TrimSpace(summary)
	modelName := strings.TrimSpace(orm.AsString(ev.Payload["model"]))
	resID, ok := orm.CoerceInt64(ev.Payload["id"])
	if xmlid == "" || modelName == "" || !ok || 0 >= resID {
		return fmt.Errorf("activity action requires type, model, and id")
	}
	if _, registered := orm.Registry["mail.activity"]; !registered {
		return fmt.Errorf("mail.activity is not installed")
	}
	typeID, typeModel, err := orm.ResolveXmlId(ctx, xmlid)
	if err != nil || typeModel != "mail.activity.type" || 0 >= typeID {
		return fmt.Errorf("unknown activity type %q", xmlid)
	}
	if summary == "" {
		summary = xmlid
	}
	vals := map[string]interface{}{
		"model":            modelName,
		"res_id":           int(resID),
		"activity_type_id": typeID,
		"summary":          summary,
		"state":            "planned",
	}
	if ev.Actor > 0 {
		vals["user_id"] = ev.Actor
	}
	_, err = orm.Create(ctx, orm.Registry["mail.activity"], vals)
	return err
}
