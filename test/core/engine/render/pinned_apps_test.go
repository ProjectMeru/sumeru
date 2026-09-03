package render_test

import (
	"testing"

	"sumeru/core/engine/render"
)

func TestParsePinnedAppsJSON(t *testing.T) {
	if got := render.ParsePinnedAppsJSON(""); got != nil {
		t.Fatalf("empty: got %v want nil", got)
	}
	if got := render.ParsePinnedAppsJSON("[]"); got != nil {
		t.Fatalf("[]: got %v want nil", got)
	}
	if got := render.ParsePinnedAppsJSON(`["mail","contacts"]`); len(got) != 2 || got[0] != "mail" {
		t.Fatalf("got %#v", got)
	}
	if got := render.ParsePinnedAppsJSON("{bad"); got != nil {
		t.Fatalf("bad json: got %v want nil", got)
	}
}

func TestSanitizePinnedModuleList(t *testing.T) {
	allowed := map[string]struct{}{
		"mail":     {},
		"contacts": {},
	}
	raw := []string{"contacts", "base", "mail", "contacts", "unknown", "  mail  "}
	got := render.SanitizePinnedModuleList(raw, allowed)
	want := []string{"contacts", "mail"}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
	if got := render.SanitizePinnedModuleList(nil, allowed); got != nil {
		t.Fatalf("nil raw: got %v", got)
	}
	if got := render.SanitizePinnedModuleList(raw, nil); got != nil {
		t.Fatalf("nil allowed: got %v", got)
	}
}
