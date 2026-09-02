package queue

import "context"

var publishMirrorHook func(ctx context.Context, topic string, payload []byte)

// SetPublishMirrorHook registers cross-instance publish forwarding (enterprise ops).
func SetPublishMirrorHook(fn func(ctx context.Context, topic string, payload []byte)) {
	publishMirrorHook = fn
}

// ClearPublishMirrorHook removes the mirror hook (tests).
func ClearPublishMirrorHook() {
	publishMirrorHook = nil
}

func publishRedisMirror(ctx context.Context, topic string, payload []byte) {
	if publishMirrorHook == nil {
		return
	}
	publishMirrorHook(ctx, topic, payload)
}
