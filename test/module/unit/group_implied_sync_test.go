package module_test

import (
	"reflect"
	"testing"

	"sumeru/core/module"
)

func TestExtractImpliedGroupXMLRefs(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single double quote", `[(4, ref("base.group_user"))]`, []string{"base.group_user"}},
		{"single single quote", `[(4, ref('sale.group_manager'))]`, []string{"sale.group_manager"}},
		{"multiple spaced", `[(4, ref("a.b")),(4, ref("c.d"))]`, []string{"a.b", "c.d"}},
		{"no match", "[(6, 0, [])]", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := module.ExtractImpliedGroupXMLRefs(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ExtractImpliedGroupXMLRefs(%q) = %#v; want %#v", tt.in, got, tt.want)
			}
		})
	}
}
