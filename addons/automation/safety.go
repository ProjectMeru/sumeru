package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/sdk"
	"sumeru/core/security"
)

const automationDepthCap = 5

type depthKey struct{}

func withDepth(ctx context.Context, n int) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, depthKey{}, n)
}

func depthOf(ctx context.Context) int {
	if ctx == nil {
		return 0
	}
	n, _ := ctx.Value(depthKey{}).(int)
	return n
}

func actorContext(ctx context.Context, actor int) context.Context {
	out := sdk.ContextWithUID(context.Background(), actor)
	if cid := sdk.CompanyIDFromContext(ctx); cid > 0 {
		out = sdk.ContextWithCompanyID(out, cid)
	}
	return withDepth(out, depthOf(ctx))
}

func automationSkippedModel(model string) bool {
	switch strings.TrimSpace(model) {
	case "mail.message", "sys.audit", "sys.server.action", "sys.server.action.log", "sys.outbox.event":
		return true
	default:
		return false
	}
}

var deleteMatches sync.Map

func deleteMatchKey(model string, id int) string {
	return model + "|" + fmt.Sprint(id)
}

func noteDeleteMatch(model string, id int, actionIDs map[int64]struct{}) {
	deleteMatches.Store(deleteMatchKey(model, id), actionIDs)
}

func deleteActionMatched(model string, id int, actionID int64) bool {
	raw, ok := deleteMatches.Load(deleteMatchKey(model, id))
	if !ok {
		return false
	}
	set, _ := raw.(map[int64]struct{})
	_, hit := set[actionID]
	return hit
}

func matchBeforeDelete(ctx context.Context, model string, id int, before map[string]interface{}) {
	if !sdk.DatabaseOpen() || 0 >= id || automationSkippedModel(model) {
		return
	}
	if !sdk.ModelRegistered("sys.server.action") {
		return
	}
	_ = sdk.WithElevated(ctx, "automation.delete_match", func(bypass context.Context) error {
		rows, err := sdk.Search(bypass, "sys.server.action", [][]interface{}{
			{"event_name", "=", sdk.EventRecordDeleted},
			{"active", "=", true},
		})
		if err != nil {
			return err
		}
		matched := map[int64]struct{}{}
		for _, row := range rows {
			actionModel := strings.TrimSpace(sdk.AsString(row["model"]))
			if actionModel != "" && actionModel != model {
				continue
			}
			domRaw := strings.TrimSpace(sdk.AsString(row["trigger_domain"]))
			if domRaw == "" {
				continue
			}
			var parsed [][]interface{}
			if err := json.Unmarshal([]byte(domRaw), &parsed); err != nil {
				applog.Warn(bypass, applog.Event{
					Message:   "server action domain JSON invalid",
					Component: "automation",
					Operation: "server_action",
					Status:    "failed",
					Context:   map[string]interface{}{"action": row["name"]},
					Err:       err,
				})
				continue
			}
			if !sdk.RecordMatchesDomain(model, before, parsed) {
				continue
			}
			actionID, ok := sdk.CoerceInt64(row["id"])
			if ok {
				matched[actionID] = struct{}{}
			}
		}
		noteDeleteMatch(model, id, matched)
		return nil
	})
}

func claimRun(ctx context.Context, actionID int64, ev event.Event, model string, resID int64) (bool, error) {
	if !sdk.DatabaseOpen() || 0 >= actionID {
		return true, nil
	}
	if !sdk.ModelRegistered("sys.server.action.log") {
		return true, nil
	}
	bucket := time.Now().UTC().Unix() / 60
	key := fmt.Sprintf("%d|%s|%s|%d|%d", actionID, ev.Name, model, resID, bucket)
	claimed := false
	err := sdk.WithElevated(ctx, "automation.action_log", func(bypass context.Context) error {
		_, err := sdk.Create(bypass, "sys.server.action.log", map[string]interface{}{
			"action_id":  actionID,
			"event_name": ev.Name,
			"model":      model,
			"res_id":     resID,
			"ran_at":     time.Now().UTC().Format(time.RFC3339),
			"dedupe_key": key,
			"detail":     "run",
		})
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
				return nil
			}
			return err
		}
		claimed = true
		return nil
	})
	return claimed, err
}

func rejectUnsafeWrite(modelName string, vals map[string]interface{}) error {
	names, ok := sdk.ModelFieldNames(modelName)
	if !ok {
		return fmt.Errorf("unknown model %q", modelName)
	}
	known := map[string]struct{}{}
	for _, field := range names {
		known[field] = struct{}{}
	}
	policies := security.FieldRegistry[modelName]
	for key := range vals {
		if _, ok := known[key]; !ok {
			return fmt.Errorf("unknown field %q on %s", key, modelName)
		}
		if pol, ok := policies[key]; ok && (pol.ReadRedact || pol.WriteDenyDirect || pol.WriteDenyUnlessSys) {
			return fmt.Errorf("field %q cannot be set by a server action", key)
		}
	}
	return nil
}
