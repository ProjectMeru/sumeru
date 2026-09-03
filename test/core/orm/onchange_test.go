package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestRegisterOnchange(t *testing.T) {
	orm.RegisterOnchange("test.model", "name", func(ctx context.Context, values map[string]interface{}, field string) (orm.OnchangeResult, error) {
		return orm.OnchangeResult{Value: map[string]interface{}{"note": "ok"}}, nil
	})
	result, err := orm.RunOnchange(context.Background(), "test.model", "name", map[string]interface{}{"name": "A"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Value["note"] != "ok" {
		t.Fatalf("expected value note=ok, got %v", result.Value)
	}
}
