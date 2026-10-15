package web

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

const maxBusSubscriptions = 64

// AuthorizeSwcBusChannel checks whether uid may subscribe to channel.
func AuthorizeSwcBusChannel(ctx context.Context, uid int, channel string) bool {
	if 0 >= uid || channel == "" {
		return false
	}
	channel = strings.TrimSpace(channel)
	parts := strings.Split(channel, "/")
	if len(parts) < 2 {
		return false
	}
	switch parts[0] {
	case "user":
		id, err := strconv.Atoi(parts[1])
		return err == nil && id == uid
	case "group":
		if len(parts) < 2 {
			return false
		}
		xmlid := strings.Join(parts[1:], ".")
		return orm.UserHasGroupXML(ctx, uid, xmlid)
	case "record":
		if len(parts) != 3 {
			return false
		}
		model := parts[1]
		rid, err := strconv.Atoi(parts[2])
		if err != nil || 0 >= rid {
			return false
		}
		if err := orm.CheckModelAccess(ctx, uid, model, "read"); err != nil {
			return false
		}
		_, err = orm.SearchOne(ctx, model, map[string]interface{}{"id": rid})
		return err == nil
	case "model":
		if len(parts) != 2 {
			return false
		}
		return orm.CheckModelAccess(ctx, uid, parts[1], "read") == nil
	case "company":
		if len(parts) != 2 {
			return false
		}
		cid, err := strconv.Atoi(parts[1])
		if err != nil || 0 >= cid {
			return false
		}
		return orm.UserAllowedCompany(ctx, uid, int64(cid))
	default:
		return false
	}
}

// RecordBusChannel builds the standard record mutation channel name.
func RecordBusChannel(model string, id int) string {
	model = strings.TrimSpace(model)
	if model == "" || 0 >= id {
		return ""
	}
	return fmt.Sprintf("record/%s/%d", model, id)
}

// UserNotificationsChannel is the inbox channel for uid.
func UserNotificationsChannel(uid int) string {
	return orm.UserNotificationsBusChannel(uid)
}
