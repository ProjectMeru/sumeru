package digest

import (
	"context"
	"fmt"
	"strings"

	"sumeru/addons/mail"
	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("cron.tick", runDueDigests)
}

func runDueDigests(ctx context.Context, _ event.Event) error {
	if orm.DB == nil {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	rows, err := orm.Search(bypass, "digest.digest", [][]interface{}{{"active", "=", true}})
	if err != nil || len(rows) == 0 {
		return err
	}
	for _, digestRow := range rows {
		if err := runDigest(bypass, digestRow); err != nil {
			applog.Warn(ctx, applog.Event{
				Message:   "digest run failed",
				Component: "digest",
				Operation: "cron.tick",
				Status:    "failed",
				Err:       err,
			})
		}
	}
	return nil
}

func runDigest(ctx context.Context, digestRow map[string]interface{}) error {
	digestID, ok := orm.CoerceInt64(digestRow["id"])
	if !ok || digestID <= 0 {
		return nil
	}
	name := strings.TrimSpace(orm.AsString(digestRow["name"]))
	if name == "" {
		name = fmt.Sprintf("digest #%d", digestID)
	}
	kpiRows, err := orm.Search(ctx, "digest.kpi", [][]interface{}{{"digest_id", "=", int(digestID)}})
	if err != nil {
		return err
	}
	if len(kpiRows) == 0 {
		return nil
	}
	var lines []string
	for _, kpi := range kpiRows {
		code := strings.TrimSpace(orm.AsString(kpi["compute_code"]))
		label := strings.TrimSpace(orm.AsString(kpi["name"]))
		if code == "" {
			continue
		}
		if label == "" {
			label = code
		}
		value, err := ComputeKPI(ctx, code)
		if err != nil {
			lines = append(lines, fmt.Sprintf("- %s: unavailable (%v)", label, err))
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s: %.0f", label, value))
	}
	if len(lines) == 0 {
		return nil
	}
	body := name + "\n" + strings.Join(lines, "\n")
	return mail.PostMessage(ctx, "digest.digest", digestID, body, mail.SubtypeNotification, "Digest")
}
