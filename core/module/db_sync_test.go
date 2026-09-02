package module

import (
	"context"
	"testing"
)

func TestResolveModuleCategoryIDEmpty(t *testing.T) {
	id, err := resolveModuleCategoryID(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if id != nil {
		t.Fatalf("expected nil category id, got %v", id)
	}
}
