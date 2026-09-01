package orm

import (
	"context"
	"fmt"
	"testing"
)

func TestBuildSearchWhereClauseComparisonOps(t *testing.T) {
	registerStubModel(t, "test.cmp.domain", []FieldDefinition{
		{Name: "amount", Type: Float},
	})

	for _, tc := range []struct {
		op   string
		val  interface{}
		want string
	}{
		{">", 10, ">"},
		{">=", 10, ">="},
		{"<", 5, "<"},
		{"<=", 5, "<="},
	} {
		where, args, err := buildSearchWhereClause("test.cmp.domain", [][]interface{}{
			{"amount", tc.op, tc.val},
		})
		if err != nil {
			t.Fatalf("op %s: %v", tc.op, err)
		}
		if len(args) != 1 || args[0] != tc.val {
			t.Fatalf("op %s args=%v", tc.op, args)
		}
		if where == "" || !contains(where, tc.want) {
			t.Fatalf("op %s where=%q want %s", tc.op, where, tc.want)
		}
	}
}

func TestValidateConstraints(t *testing.T) {
	model := "test.constraint.model"
	registerStubModel(t, model, []FieldDefinition{
		{Name: "name", Type: Char},
		{Name: "qty", Type: Integer},
	})
	RegisterConstraint(model, "qty_positive", func(ctx context.Context, rec map[string]interface{}) error {
		q, _ := CoerceInt64(rec["qty"])
		if q < 0 {
			return fmt.Errorf("qty must be non-negative")
		}
		return nil
	})
	if err := ValidateConstraints(context.Background(), model, map[string]interface{}{"qty": -1}); err == nil {
		t.Fatal("expected constraint error")
	}
	if err := ValidateConstraints(context.Background(), model, map[string]interface{}{"qty": 1}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
