package applog_test

import (
	"context"
	"errors"
	"testing"

	"sumeru/core/applog"
	"sumeru/core/server/config"
)

func TestWarnDebugErrorAndCompanyResolver(t *testing.T) {
	if err := applog.SetupFromConfig(&config.Config{
		LogEnabled: true,
		LogStdout:  false,
		LogFile:    t.TempDir() + "/app.log",
		DevMode:    true,
	}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	applog.RegisterCompanyIDResolver(func(context.Context) int64 { return 99 })
	applog.Warn(ctx, applog.Event{Component: "t", Operation: "op", Err: errors.New("warn")})
	applog.Debug(ctx, applog.Event{Component: "t", Operation: "op"})
	applog.Error(ctx, applog.Event{Component: "t", Operation: "op", Err: errors.New("fail")})
}

func TestRequestIDAndLoggerHelpers(t *testing.T) {
	id := applog.NewRequestID()
	if id == "" {
		t.Fatal("empty request id")
	}
	ctx := applog.ContextWithRequestID(context.Background(), id)
	if got := applog.RequestIDFromContext(ctx); got != id {
		t.Fatalf("got %q", got)
	}
	if got := applog.RequestIDFromContext(context.TODO()); got != "" {
		t.Fatal("empty ctx")
	}
}

func TestLoggerFromContextExtended(t *testing.T) {
	ctx := context.Background()
	logger := applog.LoggerFromContext(ctx)
	if logger == nil {
		t.Fatal("nil logger")
	}
}
