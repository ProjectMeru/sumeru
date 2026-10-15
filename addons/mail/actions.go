package mail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("mail.activity", "action_done", actionDone)
	orm.RegisterObjectAction("mail.activity", "action_snooze", actionSnooze)
	orm.RegisterSearchResultFilter(filterParentReadable)
}

func actionDone(ctx context.Context, model string, id int, _ map[string]string) (string, error) {
	if model != "mail.activity" || 0 >= id {
		return "", fmt.Errorf("record not found")
	}
	rec, err := orm.SearchOne(ctx, "mail.activity", map[string]interface{}{"id": id})
	if err != nil {
		return "", fmt.Errorf("record not found")
	}
	if strings.TrimSpace(orm.AsString(rec["state"])) == "done" {
		return "", nil
	}
	if !activityActorAllowed(ctx, rec) {
		return "", fmt.Errorf("access denied")
	}
	return "", orm.UpdateRecordByID(ctx, "mail.activity", id, map[string]interface{}{
		"state":     "done",
		"date_done": time.Now().UTC().Format("2006-01-02"),
	})
}

func actionSnooze(ctx context.Context, model string, id int, _ map[string]string) (string, error) {
	if model != "mail.activity" || 0 >= id {
		return "", fmt.Errorf("record not found")
	}
	rec, err := orm.SearchOne(ctx, "mail.activity", map[string]interface{}{"id": id})
	if err != nil {
		return "", fmt.Errorf("record not found")
	}
	if strings.TrimSpace(orm.AsString(rec["state"])) == "done" {
		return "", fmt.Errorf("done activities cannot be snoozed")
	}
	if !activityActorAllowed(ctx, rec) {
		return "", fmt.Errorf("access denied")
	}
	days := 1
	typeID := recordID(rec["activity_type_id"])
	if typeID > 0 {
		typeRow, err := orm.SearchOne(ctx, "mail.activity.type", map[string]interface{}{"id": int(typeID)})
		if err == nil {
			if n, ok := orm.CoerceInt64(typeRow["delay_count"]); ok && n > 0 {
				days = int(n)
			}
		}
	}
	base := time.Now().UTC()
	if raw := strings.TrimSpace(orm.AsString(rec["date_deadline"])); len(raw) >= 10 {
		if parsed, err := time.Parse("2006-01-02", raw[:10]); err == nil {
			base = parsed
		}
	}
	next := base.AddDate(0, 0, days).Format("2006-01-02")
	return "", orm.UpdateRecordByID(ctx, "mail.activity", id, map[string]interface{}{
		"date_deadline": next,
	})
}

func activityActorAllowed(ctx context.Context, rec map[string]interface{}) bool {
	uid := orm.SecurityUID(ctx)
	if uid > 0 && recordID(rec["user_id"]) == int64(uid) {
		return true
	}
	return orm.UserHasGroupXML(ctx, uid, "base.group_system")
}

func recordID(v interface{}) int64 {
	if n, ok := orm.CoerceInt64(v); ok {
		return n
	}
	items, ok := v.([]interface{})
	if ok && len(items) > 0 {
		n, _ := orm.CoerceInt64(items[0])
		return n
	}
	return 0
}
