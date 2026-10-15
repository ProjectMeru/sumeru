package orm

import (
	"context"
	"strconv"
)

var busPublishHook func(ctx context.Context, channel string, payload map[string]interface{})

// SetBusPublishHook registers the server implementation (WebSocket fan-out).
func SetBusPublishHook(fn func(ctx context.Context, channel string, payload map[string]interface{})) {
	busPublishHook = fn
}

// PublishBusChannel emits a bus event when a hook is registered.
func PublishBusChannel(ctx context.Context, channel string, payload map[string]interface{}) {
	if busPublishHook != nil && channel != "" {
		busPublishHook(ctx, channel, payload)
	}
}

// UserNotificationsBusChannel returns the inbox channel for uid.
func UserNotificationsBusChannel(uid int) string {
	if 0 >= uid {
		return ""
	}
	return "user/" + strconv.Itoa(uid) + "/notifications"
}
