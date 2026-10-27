package mail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/sdk"
)

func init() {
	sdk.RegisterMethod("mail.activity", "action_done", actionDone)
	sdk.RegisterMethod("mail.activity", "action_snooze", actionSnooze)
	sdk.RegisterSearchResultFilter(filterParentReadable)
}

func actionDone(ctx context.Context, rs sdk.Recordset) (sdk.ActionResult, error) {
	if _, err := rs.EnsureOne(); err != nil || rs.Model() != "mail.activity" {
		return sdk.ActionResult{}, fmt.Errorf("record not found")
	}
	rows, err := rs.Read(nil)
	if err != nil || 0 == len(rows) {
		return sdk.ActionResult{}, fmt.Errorf("record not found")
	}
	rec := rows[0]
	if strings.TrimSpace(sdk.AsString(rec["state"])) == "done" {
		return sdk.ActionResult{}, nil
	}
	if !activityActorAllowed(ctx, rec) {
		return sdk.ActionResult{}, fmt.Errorf("access denied")
	}
	_, err = rs.Write(map[string]interface{}{
		"state":     "done",
		"date_done": time.Now().UTC().Format("2006-01-02"),
	})
	return sdk.ActionResult{}, err
}

func actionSnooze(ctx context.Context, rs sdk.Recordset) (sdk.ActionResult, error) {
	if _, err := rs.EnsureOne(); err != nil || rs.Model() != "mail.activity" {
		return sdk.ActionResult{}, fmt.Errorf("record not found")
	}
	rows, err := rs.Read(nil)
	if err != nil || 0 == len(rows) {
		return sdk.ActionResult{}, fmt.Errorf("record not found")
	}
	rec := rows[0]
	if strings.TrimSpace(sdk.AsString(rec["state"])) == "done" {
		return sdk.ActionResult{}, fmt.Errorf("done activities cannot be snoozed")
	}
	if !activityActorAllowed(ctx, rec) {
		return sdk.ActionResult{}, fmt.Errorf("access denied")
	}
	days := 1
	typeID := recordID(rec["activity_type_id"])
	if typeID > 0 {
		typeRow, err := sdk.SearchOne(ctx, "mail.activity.type", map[string]interface{}{"id": int(typeID)})
		if err == nil {
			if n, ok := sdk.CoerceInt64(typeRow["delay_count"]); ok && n > 0 {
				days = int(n)
			}
		}
	}
	base := time.Now().UTC()
	if raw := strings.TrimSpace(sdk.AsString(rec["date_deadline"])); len(raw) >= 10 {
		if parsed, err := time.Parse("2006-01-02", raw[:10]); err == nil {
			base = parsed
		}
	}
	next := base.AddDate(0, 0, days).Format("2006-01-02")
	_, err = rs.Write(map[string]interface{}{"date_deadline": next})
	return sdk.ActionResult{}, err
}

func activityActorAllowed(ctx context.Context, rec map[string]interface{}) bool {
	uid := sdk.SecurityUID(ctx)
	if uid > 0 && recordID(rec["user_id"]) == int64(uid) {
		return true
	}
	return sdk.UserHasGroupXML(ctx, uid, "base.group_system")
}

func recordID(v interface{}) int64 {
	if n, ok := sdk.CoerceInt64(v); ok {
		return n
	}
	items, ok := v.([]interface{})
	if ok && len(items) > 0 {
		n, _ := sdk.CoerceInt64(items[0])
		return n
	}
	return 0
}
