package module_test

import (
	"context"
	"testing"

	"sumeru/core/module"
)

func TestResolveModuleCategoryIDEmpty(t *testing.T) {
	id, err := module.ResolveModuleCategoryIDForTest(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if id != nil {
		t.Fatalf("expected nil category id, got %v", id)
	}
}
