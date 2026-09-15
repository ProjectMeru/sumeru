package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestRegisterOnchange(t *testing.T) {
	orm.RegisterOnchange("core.partner", "name", func(ctx context.Context, values map[string]interface{}, field string) (orm.OnchangeResult, error) {
		return orm.OnchangeResult{Value: map[string]interface{}{"note": "ok"}}, nil
	})
	ctx := orm.ContextWithUID(context.Background(), 1)
	result, err := orm.RunOnchange(ctx, "core.partner", "name", map[string]interface{}{"name": "A"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Value["note"] != "ok" {
		t.Fatalf("expected value note=ok, got %v", result.Value)
	}
}
