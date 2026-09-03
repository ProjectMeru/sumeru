package module_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/module"
)

func TestUpdateModuleData_unknownModule(t *testing.T) {
	err := module.UpdateModuleData(context.Background(), "no_such_module_phase3")
	if err == nil || !strings.Contains(err.Error(), "unknown module") {
		t.Fatalf("UpdateModuleData() = %v; want unknown module error", err)
	}
}

func TestModuleViewDeleteQuerySkipsSharedCoreIDs(t *testing.T) {
	q := module.ModuleViewDeleteQuery(`"sys_view"`, `"sys_model_data"`)
	if !strings.Contains(q, "NOT EXISTS") {
		t.Fatal("view delete query must exclude core_id shared with other modules")
	}
	if !strings.Contains(q, "other.module <> $1") {
		t.Fatal("view delete query must compare module names")
	}
}
