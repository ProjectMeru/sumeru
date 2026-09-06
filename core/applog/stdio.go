package applog

import "context"

func InfoMsg(ctx context.Context, component, operation, message string, ctxFields map[string]interface{}) {
	Info(ctx, Event{
		Message: message, Component: component, Operation: operation,
		Status: "success", Context: ctxFields,
	})
}

func WarnMsg(ctx context.Context, component, operation, message string, err error, ctxFields map[string]interface{}) {
	Warn(ctx, Event{
		Message: message, Component: component, Operation: operation,
		Status: "partial", Context: ctxFields, Err: err,
	})
}

func DebugMsg(ctx context.Context, component, operation, message string, ctxFields map[string]interface{}) {
	Debug(ctx, Event{
		Message: message, Component: component, Operation: operation,
		Status: "success", Context: ctxFields,
	})
}

func ErrorCode(ctx context.Context, code, message string, ev Event) {
	ev.Code = code
	if message != "" {
		ev.Message = message
	}
	if ev.Status == "" {
		ev.Status = "failure"
	}
	Error(ctx, ev)
}

func WarnCode(ctx context.Context, code, message string, ev Event) {
	ev.Code = code
	if message != "" {
		ev.Message = message
	}
	if ev.Status == "" {
		ev.Status = "partial"
	}
	Warn(ctx, ev)
}
