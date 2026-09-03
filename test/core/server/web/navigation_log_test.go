package web_test

import (
	"context"
	"testing"

	"sumeru/core/applog"
	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func TestWebLogNavigation_emitsWithoutPanic(t *testing.T) {
	cfg := &config.Config{LogEnabled: true, LogStdout: true, DevMode: false}
	if err := applog.SetupFromConfig(cfg); err != nil {
		t.Fatal(err)
	}
	defer applog.Sync()

	ctx := context.Background()
	web.WebLogNavigation(ctx, "/web", "view_open", "Workspace view opened", map[string]interface{}{
		"menu_id":   "12",
		"action_id": 3,
		"model":     "core.user",
		"view_type": "form",
	})
}

func TestSetupFromConfig_requiresLogSink(t *testing.T) {
	cfg := &config.Config{LogEnabled: true, LogStdout: false, LogFile: ""}
	if err := applog.SetupFromConfig(cfg); err == nil {
		t.Fatal("expected error when log_stdout and log_file are both unset")
	}
}
