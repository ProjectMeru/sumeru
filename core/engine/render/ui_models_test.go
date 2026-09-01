package render

import "testing"

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
		if got := UIModelName(tt.model); got != tt.want {
			t.Errorf("UIModelName(%q) = %q, want %q", tt.model, got, tt.want)
		}
	}
}

func TestHumanViewBreadcrumb(t *testing.T) {
	tests := []struct {
		model, viewType, want string
	}{
		{"core.company", ViewModeList, "Companies"},
		{"core.company", ViewModeForm, "Company"},
		{"core.user", ViewModeKanban, "User"},
		{"core.group", ViewModePivot, "Groups · Pivot"},
		{"core.group", ViewModeForm, "Groups"},
	}
	for _, tt := range tests {
		if got := HumanViewBreadcrumb(tt.model, tt.viewType); got != tt.want {
			t.Errorf("HumanViewBreadcrumb(%q, %q) = %q, want %q", tt.model, tt.viewType, got, tt.want)
		}
	}
}
