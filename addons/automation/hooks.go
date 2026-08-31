package automation

import (
	"context"

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
	event.Subscribe("cron.tick", runServerActionsForEvent)
}

func runServerActionsForEvent(ctx context.Context, ev event.Event) error {
	if orm.DB == nil {
		return nil
	}
	if _, ok := orm.Registry["sys.server.action"]; !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
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
	for _, row := range rows {
		if err := executeServerAction(ctx, row, ev); err != nil {
			applog.Warn(ctx, applog.Event{
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
}
