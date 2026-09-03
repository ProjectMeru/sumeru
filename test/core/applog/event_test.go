package applog_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sumeru/core/applog"
	"sumeru/core/server/config"
)

func TestEvent_JSONShape(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	if err := applog.SetupFromConfig(&config.Config{
		LogEnabled:  true,
		LogStdout:   false,
		LogFile:     logPath,
		LogRolling:  false,
		DevMode:     true,
		LogTimezone: "UTC",
	}); err != nil {
		t.Fatal(err)
	}

	ctx := applog.ContextWithRequestID(context.Background(), "req_test123")
	applog.RegisterUIDResolver(func(context.Context) int { return 42 })

	applog.Info(ctx, applog.Event{
		Message:   "Record created successfully",
		Component: "orm",
		Module:    "account",
		Operation: "create",
		Status:    "success",
		Duration:  18 * time.Millisecond,
		Context: map[string]interface{}{
			"resource":    "account.move",
			"resource_id": 18293,
		},
	})

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		t.Fatal("expected log output")
	}
	line := lines[len(lines)-1]
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, line)
	}
	if _, ok := m["time"]; !ok {
		t.Fatal("missing time")
	}
	if _, ok := m["log_ts"]; ok {
		t.Fatal("log_ts must not be duplicated")
	}
	if _, ok := m["source"]; ok {
		t.Fatal("source must not appear in log output")
	}
	if m["message"] != "Record created successfully" {
		t.Fatalf("message=%v", m["message"])
	}
	if m["component"] != "orm" {
		t.Fatalf("component=%v", m["component"])
	}
	if m["request_id"] != "req_test123" {
		t.Fatalf("request_id=%v", m["request_id"])
	}
	ctxObj, ok := m["context"].(map[string]interface{})
	if !ok {
		t.Fatalf("context=%T", m["context"])
	}
	if ctxObj["resource"] != "account.move" {
		t.Fatalf("resource=%v", ctxObj["resource"])
	}
}
