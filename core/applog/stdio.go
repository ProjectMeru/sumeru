package applog

import "context"

// InfoMsg logs a structured info event with the logging contract.
func InfoMsg(ctx context.Context, component, operation, message string, ctxFields map[string]interface{}) {
	Info(ctx, Event{
		Message:   message,
		Component: component,
		Operation: operation,
		Status:    "success",
		Context:   ctxFields,
	})
}

// WarnMsg logs a structured warning event with the logging contract.
func WarnMsg(ctx context.Context, component, operation, message string, err error, ctxFields map[string]interface{}) {
	Warn(ctx, Event{
		Message:   message,
		Component: component,
		Operation: operation,
		Status:    "partial",
		Context:   ctxFields,
		Err:       err,
	})
}

// DebugMsg logs a structured debug event with the logging contract.
func DebugMsg(ctx context.Context, component, operation, message string, ctxFields map[string]interface{}) {
	Debug(ctx, Event{
		Message:   message,
		Component: component,
		Operation: operation,
		Status:    "success",
		Context:   ctxFields,
	})
}

// ErrorCode logs a failure with a stable machine code and human message.
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

// WarnCode logs a warning with a stable machine code and human message.
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
