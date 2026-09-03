package render_test

import (
	"testing"
	"sumeru/core/engine/render"
)


func TestUserInitialsFromName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "?"},
		{"  ", "?"},
		{"Ada", "AD"},
		{"A", "A"},
		{"Jean Dupont", "JD"},
		{"Mary Jane Watson", "MW"},
	}
	for _, tt := range tests {
		if got := render.UserInitialsFromName(tt.in); got != tt.want {
			t.Errorf("UserInitialsFromName(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
