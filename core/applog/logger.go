package applog

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"sumeru/core/errcode"
)

var (
	uidMu       sync.RWMutex
	uidFromCtx  func(context.Context) int
	logLocation *time.Location
	logEnabled  bool
	logTzName   string
)

// RegisterUIDResolver wires user_id resolution for event context maps.
func RegisterUIDResolver(fn func(context.Context) int) {
	uidMu.Lock()
	defer uidMu.Unlock()
	uidFromCtx = fn
}

func resolveUID(ctx context.Context) int {
	uidMu.RLock()
	fn := uidFromCtx
	uidMu.RUnlock()
	if fn == nil || ctx == nil {
		return 0
	}
	return fn(ctx)
}

func baseLogger() *slog.Logger {
	if root != nil {
		return root
	}
	return slog.Default()
}

// LoggerFromContext returns the base logger. Per-event fields use the Event API;
// slog JSON handler supplies the canonical top-level "time" field (no log_ts duplication).
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if !logEnabled {
		return slog.New(slog.DiscardHandler)
	}
	return baseLogger()
}

// L is an alias for LoggerFromContext.
func L(ctx context.Context) *slog.Logger {
	return LoggerFromContext(ctx)
}

// Fatal logs at error level and exits the process.
func Fatal(ctx context.Context, msg string, keysAndValues ...interface{}) {
	ev := Event{
		Message:   msg,
		Code:      errcode.InternalError,
		Component: "server",
		Status:    "failure",
	}
	if len(keysAndValues) > 0 {
		ev.Context = kvPairsToMap(keysAndValues)
	}
	Error(ctx, ev)
	os.Exit(1)
}

func kvPairsToMap(pairs []interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for i := 0; i+1 < len(pairs); i += 2 {
		if k, ok := pairs[i].(string); ok {
			if e, ok := pairs[i+1].(error); ok {
				if e != nil {
					out[k] = e.Error()
				}
			} else {
				out[k] = pairs[i+1]
			}
		}
	}
	return out
}
