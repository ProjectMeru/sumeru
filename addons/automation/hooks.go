package automation

import (
	"context"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	registerHooks()
}

func registerHooks() {
	event.Subscribe("record.created", runServerActionsForEvent)
	event.Subscribe("record.updated", runServerActionsForEvent)
	event.Subscribe("record.deleted", runServerActionsForEvent)
	event.Subscribe("cron.tick", runServerActionsForEvent)
	orm.RegisterBeforeDelete(matchBeforeDelete)
}

func runServerActionsForEvent(ctx context.Context, ev event.Event) error {
	if orm.DB == nil {
		return nil
	}
	if _, ok := orm.Registry["sys.server.action"]; !ok {
		return nil
	}
	modelName := strings.TrimSpace(orm.AsString(ev.Payload["model"]))
	if automationSkippedModel(modelName) {
		return nil
	}
	return orm.WithElevated(ctx, "automation.server_action", func(bypass context.Context) error {
		installed, err := orm.InstalledModuleNames(bypass)
		if err != nil || !orm.ShouldMaterializeModel("sys.server.action", installed) {
			return nil
		}
		rows, err := orm.Search(bypass, "sys.server.action", [][]interface{}{
			{"event_name", "=", ev.Name},
			{"active", "=", true},
		})
		if err != nil {
			return err
		}
		runCtx := actorContext(ctx, ev.Actor)
		for _, row := range rows {
			if !actionApplies(runCtx, row, ev) {
				continue
			}
			actionID, _ := orm.CoerceInt64(row["id"])
			resID, _ := orm.CoerceInt64(ev.Payload["id"])
			claimed, err := claimRun(runCtx, actionID, ev, modelName, resID)
			if err != nil {
				applog.Warn(runCtx, applog.Event{
					Message:   "server action log failed",
					Component: "automation",
					Operation: "server_action",
					Status:    "failed",
					Err:       err,
				})
				continue
			}
			if !claimed {
				continue
			}
			if err := executeServerAction(runCtx, row, ev); err != nil {
				applog.Warn(runCtx, applog.Event{
					Message:   "server action failed",
					Component: "automation",
					Operation: "server_action",
					Status:    "failed",
					Context:   map[string]interface{}{"event": ev.Name, "action": row["name"]},
					Err:       err,
				})
			}
		}
		return nil
	})
}
