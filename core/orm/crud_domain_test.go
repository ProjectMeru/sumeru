package orm

import (
	"strings"
	"testing"
)

func TestBuildSearchWhereClauseDateNotFalseIsNotNull(t *testing.T) {
	registerStubModel(t, "test.date.domain", []FieldDefinition{
		{Name: "date_deadline", Type: Date},
		{Name: "name", Type: Char},
	})

	where, args, err := buildSearchWhereClause("test.date.domain", [][]interface{}{
		{"date_deadline", "!=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if where == "" || !strings.Contains(where, "IS NOT NULL") {
		t.Fatalf("expected IS NOT NULL, got %q", where)
	}
}

func TestBuildSearchWhereClauseDateEqualsFalseIsNull(t *testing.T) {
	registerStubModel(t, "test.date.domain.eq", []FieldDefinition{
		{Name: "date_deadline", Type: Date},
	})

	where, args, err := buildSearchWhereClause("test.date.domain.eq", [][]interface{}{
		{"date_deadline", "=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if where == "" || !strings.Contains(where, "IS NULL") {
		t.Fatalf("expected IS NULL, got %q", where)
	}
}

func TestBuildSearchWhereClauseNonDateFalseUsesDistinctFrom(t *testing.T) {
	registerStubModel(t, "test.bool.domain", []FieldDefinition{
		{Name: "active", Type: Boolean},
	})

	where, args, err := buildSearchWhereClause("test.bool.domain", [][]interface{}{
		{"active", "!=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 1 || args[0] != false {
		t.Fatalf("expected one false arg, got args=%v", args)
	}
	if strings.Contains(where, "IS NOT NULL") {
		t.Fatalf("boolean field should not use IS NOT NULL, got %q", where)
	}
	if !strings.Contains(where, "IS DISTINCT FROM") {
		t.Fatalf("expected IS DISTINCT FROM, got %q", where)
	}
}

func TestSplitDomainORPrefix(t *testing.T) {
	orCount, leaves, ok := splitDomainORPrefix([][]interface{}{
		{"|"},
		{"name", "=", "a"},
		{"name", "=", "b"},
	})
	if !ok || orCount != 1 || len(leaves) != 2 {
		t.Fatalf("valid OR prefix: orCount=%d leaves=%d ok=%v", orCount, len(leaves), ok)
	}

	orCount, _, ok = splitDomainORPrefix([][]interface{}{
		{"|"},
		{"|"},
		{"a", "=", 1},
		{"b", "=", 2},
		{"c", "=", 3},
	})
	if !ok || orCount != 2 {
		t.Fatalf("two OR markers: orCount=%d ok=%v", orCount, ok)
	}

	_, _, ok = splitDomainORPrefix([][]interface{}{
		{"|"},
		{"name", "=", "a"},
	})
	if ok {
		t.Fatal("expected invalid shape for mismatched leaf count")
	}

	orCount, _, ok = splitDomainORPrefix([][]interface{}{{"name", "=", "a"}})
	if !ok || orCount != 0 {
		t.Fatalf("non-OR domain: orCount=%d ok=%v", orCount, ok)
	}
}

func TestBuildSearchWhereClauseORPrefix(t *testing.T) {
	registerStubModel(t, "test.or.domain", []FieldDefinition{
		{Name: "name", Type: Char},
	})

	where, args, err := buildSearchWhereClause("test.or.domain", [][]interface{}{
		{"|"},
		{"name", "=", "a"},
		{"name", "=", "b"},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if !strings.Contains(where, " OR ") {
		t.Fatalf("expected OR clause, got %q", where)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %v", args)
	}
}

func TestBuildSearchWhereClauseEmptyInAndNotIn(t *testing.T) {
	registerStubModel(t, "test.list.domain", []FieldDefinition{
		{Name: "state", Type: Selection},
	})

	where, args, err := buildSearchWhereClause("test.list.domain", [][]interface{}{
		{"state", "in", []interface{}{}},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause in: %v", err)
	}
	if where != "FALSE" || len(args) != 0 {
		t.Fatalf("empty IN: where=%q args=%v", where, args)
	}

	where, args, err = buildSearchWhereClause("test.list.domain", [][]interface{}{
		{"state", "not in", []interface{}{}},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause not in: %v", err)
	}
	if where != "TRUE" || len(args) != 0 {
		t.Fatalf("empty NOT IN: where=%q args=%v", where, args)
	}
}

func TestBuildAndWhereClauses(t *testing.T) {
	registerStubModel(t, "test.and.domain", []FieldDefinition{
		{Name: "name", Type: Char},
		{Name: "active", Type: Boolean},
	})

	where, args, err := buildAndWhereClauses("test.and.domain", [][][]interface{}{
		{{"name", "=", "a"}},
		{{"active", "=", true}},
	})
	if err != nil {
		t.Fatalf("buildAndWhereClauses: %v", err)
	}
	if !strings.Contains(where, " AND ") {
		t.Fatalf("expected AND join, got %q", where)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %v", args)
	}
}

func TestRecordMatchesDomainORPrefix(t *testing.T) {
	domain := [][]interface{}{
		{"|"},
		{"name", "=", "a"},
		{"name", "=", "b"},
	}
	if !RecordMatchesDomain(map[string]interface{}{"name": "a"}, domain) {
		t.Fatal("expected match on first leaf")
	}
	if !RecordMatchesDomain(map[string]interface{}{"name": "b"}, domain) {
		t.Fatal("expected match on second leaf")
	}
	if RecordMatchesDomain(map[string]interface{}{"name": "c"}, domain) {
		t.Fatal("expected no match")
	}
}
