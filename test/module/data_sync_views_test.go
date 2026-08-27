package module_test

import (
	"testing"

	"sumeru/core/module"
)

func TestInferSysViewTypeFromArch(t *testing.T) {
	tests := []struct {
		arch string
		want string
	}{
		{"<list string=\"X\"><field name=\"a\"/></list>", "list"},
		{"  <form><sheet/></form>", "form"},
		{"<kanban><field name=\"n\"/></kanban>", "kanban"},
		{"<search><filter name=\"a\" domain='[]'/></search>", "search"},
		{"<graph type=\"bar\"><field name=\"n\"/></graph>", "graph"},
		{"<calendar date_start=\"d\"><field name=\"n\"/></calendar>", "calendar"},
		{"<pivot><field name=\"n\" type=\"measure\"/></pivot>", "pivot"},
		{"<view type=\"list\" model=\"m\"><list/></view>", "list"},
		{"", ""},
		{"<unknown/>", ""},
	}
	for _, tt := range tests {
		if got := module.InferSysViewTypeFromArch(tt.arch); got != tt.want {
			t.Errorf("InferSysViewTypeFromArch(%q) = %q; want %q", tt.arch, got, tt.want)
		}
	}
}
