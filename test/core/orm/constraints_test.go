package orm_test

import (
	"sumeru/core/orm"
	"context"
	"fmt"
	"testing"
)



func TestBuildSearchWhereClauseComparisonOps(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.cmp.domain", []orm.FieldDefinition{
		{Name: "amount", Type: orm.Float},
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
		where, args, err := orm.BuildSearchWhereClauseForTest("test.cmp.domain", [][]interface{}{
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
	orm.RegisterStubModelForTest(t, model, []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "qty", Type: orm.Integer},
	})
	orm.RegisterConstraint(model, "qty_positive", func(ctx context.Context, rec map[string]interface{}) error {
		q, _ := orm.CoerceInt64(rec["qty"])
		if q < 0 {
			return fmt.Errorf("qty must be non-negative")
		}
		return nil
	})
	if err := orm.ValidateConstraints(context.Background(), model, map[string]interface{}{"qty": -1}); err == nil {
		t.Fatal("expected constraint error")
	}
	if err := orm.ValidateConstraints(context.Background(), model, map[string]interface{}{"qty": 1}); err != nil {
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
