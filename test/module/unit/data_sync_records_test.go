package module_test

import (
	"context"
	"reflect"
	"testing"

	"sumeru/core/module"
)

func TestConvertRecordScalar(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name   string
		column string
		raw    string
		want   interface{}
	}{
		{"perm true string", "perm_read", "true", true},
		{"perm false string", "perm_write", "false", false},
		{"perm one", "perm_create", "1", true},
		{"perm zero", "perm_unlink", "0", false},
		{"group_id empty", "group_id", "", nil},
		{"group_id false", "group_id", "false", nil},
		{"plain string", "name", "  hello  ", "hello"},
		{"active bool", "active", "true", true},
		{"suffix active", "is_active", "0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := module.ConvertRecordScalar(ctx, "base", "sys.access", tt.column, tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ConvertRecordScalar(%q, %q) = %#v; want %#v", tt.column, tt.raw, got, tt.want)
			}
		})
	}
}
