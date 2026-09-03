package render_test

import (
	"testing"

	"sumeru/core/engine/render"
)



func TestUIModelName(t *testing.T) {
	tests := []struct {
		model string
		want  string
	}{
		{"core.company", "Companies"},
		{"core.unknown", "Unknown"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := render.UIModelName(tt.model); got != tt.want {
			t.Errorf("render.UIModelName(%q) = %q, want %q", tt.model, got, tt.want)
		}
	}
}

func TestHumanViewBreadcrumb(t *testing.T) {
	tests := []struct {
		model, viewType, want string
	}{
		{"core.company", render.ViewModeList, "Companies"},
		{"core.company", render.ViewModeForm, "Company"},
		{"core.user", render.ViewModeKanban, "User"},
		{"core.group", render.ViewModePivot, "Groups · Pivot"},
		{"core.group", render.ViewModeForm, "Groups"},
	}
	for _, tt := range tests {
		if got := render.HumanViewBreadcrumb(tt.model, tt.viewType); got != tt.want {
			t.Errorf("render.HumanViewBreadcrumb(%q, %q) = %q, want %q", tt.model, tt.viewType, got, tt.want)
		}
	}
}
