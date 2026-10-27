package automation

import (
	"context"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/sdk"
)

func init() {
	registerHooks()
}

func registerHooks() {
	event.Subscribe("record.created", runServerActionsForEvent)
	event.Subscribe("record.updated", runServerActionsForEvent)
	event.Subscribe("record.deleted", runServerActionsForEvent)
	event.Subscribe("cron.tick", runServerActionsForEvent)
	sdk.RegisterBeforeDelete(matchBeforeDelete)
}

func runServerActionsForEvent(ctx context.Context, ev event.Event) error {
	if !sdk.DatabaseOpen() {
		return nil
	}
	if !sdk.ModelRegistered("sys.server.action") {
		return nil
	}
	modelName := strings.TrimSpace(sdk.AsString(ev.Payload["model"]))
	if automationSkippedModel(modelName) {
		return nil
	}
	return sdk.WithElevated(ctx, "automation.server_action", func(bypass context.Context) error {
		installed, err := sdk.InstalledModuleNames(bypass)
		if err != nil || !sdk.ShouldMaterializeModel("sys.server.action", installed) {
			return nil
		}
		rows, err := sdk.Search(bypass, "sys.server.action", [][]interface{}{
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
			actionID, _ := sdk.CoerceInt64(row["id"])
			resID, _ := sdk.CoerceInt64(ev.Payload["id"])
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
