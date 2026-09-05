package applog_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sumeru/core/applog"
	"sumeru/core/server/config"
)

func TestScrubMap_redactsSecrets(t *testing.T) {
	got := applog.ScrubMap(map[string]interface{}{
		"user_id":       1,
		"password":      "secret123",
		"api_token":     "tok",
		"Authorization": "Bearer x",
		"field":         "email",
		"nested": map[string]interface{}{
			"key_hash": "abc",
			"ok":       true,
		},
	})
	if got["password"] != applog.RedactedPlaceholder {
		t.Fatalf("password=%v", got["password"])
	}
	if got["api_token"] != applog.RedactedPlaceholder {
		t.Fatalf("api_token=%v", got["api_token"])
	}
	if got["Authorization"] != applog.RedactedPlaceholder {
		t.Fatalf("Authorization=%v", got["Authorization"])
	}
	if got["user_id"] != 1 {
		t.Fatalf("user_id=%v", got["user_id"])
	}
	nested, ok := got["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("nested=%T", got["nested"])
	}
	if nested["key_hash"] != applog.RedactedPlaceholder || nested["ok"] != true {
		t.Fatalf("nested=%v", nested)
	}
}

func TestIsSecretKey_exactSidDoesNotMatchInside(t *testing.T) {
	if !applog.IsSecretKey("sid") {
		t.Fatal("sid should be secret")
	}
	if applog.IsSecretKey("inside") {
		t.Fatal("inside must not match sid exact rule")
	}
	if !applog.TextContainsSecretKeyword(`UPDATE sys_session SET sid = $1`) {
		t.Fatal("SQL with sid column should look sensitive")
	}
}

func TestErrorCode_emitsTopLevelErrorCode(t *testing.T) {
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

	applog.ErrorCode(context.Background(), "EMAIL_ALREADY_EXISTS", "Email is already registered", applog.Event{
		Component: "web",
		Operation: "user_create",
		Context: map[string]interface{}{
			"user_id":  1,
			"field":    "email",
			"password": "should-not-appear",
		},
	})

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	line := lines[len(lines)-1]
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, line)
	}
	if m["error_code"] != "EMAIL_ALREADY_EXISTS" {
		t.Fatalf("error_code=%v", m["error_code"])
	}
	if m["message"] != "Email is already registered" {
		t.Fatalf("message=%v", m["message"])
	}
	if m["level"] != "ERROR" {
		t.Fatalf("level=%v", m["level"])
	}
	ctxObj, ok := m["context"].(map[string]interface{})
	if !ok {
		t.Fatalf("context=%T", m["context"])
	}
	if ctxObj["error_code"] != "EMAIL_ALREADY_EXISTS" {
		t.Fatalf("context.error_code=%v", ctxObj["error_code"])
	}
	if ctxObj["password"] != applog.RedactedPlaceholder {
		t.Fatalf("password not scrubbed: %v", ctxObj["password"])
	}
	if strings.Contains(line, "should-not-appear") {
		t.Fatal("secret value leaked into log line")
	}
}
