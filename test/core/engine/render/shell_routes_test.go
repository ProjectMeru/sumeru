package render_test

import (
	"sumeru/core/engine/render"
	"testing"
)

func TestModuleWebURL(t *testing.T) {
	if got := render.ModuleWebURL("crm"); got != "/web/crm" {
		t.Fatalf("got %q", got)
	}
	if got := render.ModuleWebURL("engagement_cookbook"); got != "/web/engagement_cookbook" {
		t.Fatalf("got %q", got)
	}
	if render.ModuleWebURL("apps") != "" {
		t.Fatal("reserved segment apps should be rejected")
	}
	if render.ModuleWebURL("home") != "" {
		t.Fatal("reserved segment home should be rejected")
	}
}

func TestParseDebugQuery(t *testing.T) {
	enabled, mod := render.ParseDebugQuery("true")
	if !enabled || mod != "" {
		t.Fatalf("got enabled=%v mod=%q", enabled, mod)
	}
	enabled, mod = render.ParseDebugQuery("true/crm")
	if !enabled || mod != "crm" {
		t.Fatalf("got enabled=%v mod=%q", enabled, mod)
	}
	if enabled, _ := render.ParseDebugQuery("false"); enabled {
		t.Fatal("false should not enable debug")
	}
	enabled, mod = render.ParseDebugQuery("true/not-a-valid!!!")
	if !enabled || mod != "" {
		t.Fatalf("invalid module suffix: enabled=%v mod=%q", enabled, mod)
	}
}
