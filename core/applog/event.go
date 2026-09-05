package applog

import (
	"context"
	"log/slog"
	"time"
)

// Event is the structured logging contract for all Sumeru components.
// Top-level fields are universal; module-specific data belongs in Context.
type Event struct {
	Message   string
	Code      string // stable machine id (SCREAMING_SNAKE); emitted as error_code
	Component string
	Module    string
	Operation string
	Status    string
	Duration  time.Duration
	Context   map[string]interface{}
	Err       error
}

var companyIDFromCtx func(context.Context) int64

// RegisterCompanyIDResolver wires company_id into event context maps.
func RegisterCompanyIDResolver(fn func(context.Context) int64) {
	companyIDFromCtx = fn
}

func resolveCompanyID(ctx context.Context) int64 {
	if companyIDFromCtx == nil || ctx == nil {
		return 0
	}
	return companyIDFromCtx(ctx)
}

func mergeEventContext(ctx context.Context, ev Event) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range ev.Context {
		out[k] = v
	}
	if uid := resolveUID(ctx); uid > 0 {
		if _, ok := out["user_id"]; !ok {
			out["user_id"] = uid
		}
	}
	if cid := resolveCompanyID(ctx); cid > 0 {
		if _, ok := out["company_id"]; !ok {
			out["company_id"] = cid
		}
	}
	if ev.Err != nil {
		out["error"] = ev.Err.Error()
	}
	if ev.Code != "" {
		if _, ok := out["error_code"]; !ok {
			out["error_code"] = ev.Code
		}
	}
	return ScrubMap(out)
}

func emit(ctx context.Context, level slog.Level, ev Event) {
	if !logEnabled {
		return
	}
	if ev.Message == "" {
		ev.Message = "operation completed"
	}
	attrs := []any{
		slog.String("request_id", RequestIDFromContext(ctx)),
		slog.String("component", ev.Component),
		slog.String("operation", ev.Operation),
		slog.String("status", ev.Status),
	}
	if ev.Code != "" {
		attrs = append(attrs, slog.String("error_code", ev.Code))
	}
	if ev.Module != "" {
		attrs = append(attrs, slog.String("module", ev.Module))
	}
	if ev.Duration > 0 {
		attrs = append(attrs, slog.Int64("duration_ms", ev.Duration.Milliseconds()))
	}
	ctxMap := mergeEventContext(ctx, ev)
	if len(ctxMap) > 0 {
		attrs = append(attrs, slog.Any("context", ctxMap))
	}
	baseLogger().Log(ctx, level, ev.Message, attrs...)
}

// Info logs a structured info event.
func Info(ctx context.Context, ev Event) {
	emit(ctx, slog.LevelInfo, ev)
}

// Warn logs a structured warning event.
func Warn(ctx context.Context, ev Event) {
	if ev.Status == "" {
		ev.Status = "partial"
	}
	emit(ctx, slog.LevelWarn, ev)
}

// Debug logs a structured debug event.
func Debug(ctx context.Context, ev Event) {
	emit(ctx, slog.LevelDebug, ev)
}

// Error logs a structured error event.
func Error(ctx context.Context, ev Event) {
	if ev.Status == "" {
		ev.Status = "failure"
	}
	emit(ctx, slog.LevelError, ev)
}
