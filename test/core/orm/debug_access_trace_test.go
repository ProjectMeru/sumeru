package orm_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/orm"
)

func TestBuildDebugAccessTrace_requiresLogin(t *testing.T) {
	_, err := orm.BuildDebugAccessTrace(context.Background(), 0, "core.partner")
	if err == nil || !strings.Contains(err.Error(), "login") {
		t.Fatalf("expected login error, got %v", err)
	}
}

func TestBuildDebugAccessTrace_requiresModel(t *testing.T) {
	_, err := orm.BuildDebugAccessTrace(context.Background(), 2, "  ")
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Fatalf("expected model error, got %v", err)
	}
}
