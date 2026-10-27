package digest

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/sdk"
)

func init() {
	event.Subscribe("cron.tick", runDueDigests)
}

func runDueDigests(ctx context.Context, _ event.Event) error {
	if !sdk.DatabaseOpen() {
		return nil
	}
	return sdk.WithElevated(ctx, "digest.hook", func(bypass context.Context) error {
		rows, err := sdk.Search(bypass, "digest.digest", [][]interface{}{{"active", "=", true}})
		if err != nil {
			return err
		}
		if 0 == len(rows) {
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
	})
}

func runDigest(ctx context.Context, digestRow map[string]interface{}) error {
	digestID, ok := sdk.CoerceInt64(digestRow["id"])
	if !ok || 0 >= digestID {
		return nil
	}
	name := strings.TrimSpace(sdk.AsString(digestRow["name"]))
	if name == "" {
		name = fmt.Sprintf("digest #%d", digestID)
	}
	kpiRows, err := sdk.Search(ctx, "digest.kpi", [][]interface{}{{"digest_id", "=", int(digestID)}})
	if err != nil {
		return err
	}
	if 0 == len(kpiRows) {
		return nil
	}
	var lines []string
	for _, kpi := range kpiRows {
		code := strings.TrimSpace(sdk.AsString(kpi["compute_code"]))
		label := strings.TrimSpace(sdk.AsString(kpi["name"]))
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
	if 0 == len(lines) {
		return nil
	}
	body := name + "\n" + strings.Join(lines, "\n")
	return sdk.AppendAppLog(ctx, "digest", "KPI digest", body)
}
