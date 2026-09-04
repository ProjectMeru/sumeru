package orm_test

import (
	"strings"
	"testing"

	"sumeru/core/orm"
)

func TestBuildSearchWhereClauseDateNotFalseIsNotNull(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.date.domain", []orm.FieldDefinition{
		{Name: "date_deadline", Type: orm.Date},
		{Name: "name", Type: orm.Char},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.date.domain", [][]interface{}{
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
	orm.RegisterStubModelForTest(t, "test.date.domain.eq", []orm.FieldDefinition{
		{Name: "date_deadline", Type: orm.Date},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.date.domain.eq", [][]interface{}{
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

func TestBuildSearchWhereClauseMany2OneNotFalseIsNotNull(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.m2o.domain", []orm.FieldDefinition{
		{Name: "team_id", Type: orm.Many2One, Relation: "crm.team"},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.m2o.domain", [][]interface{}{
		{"team_id", "!=", false},
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

func TestBuildSearchWhereClauseMany2OneEqualsFalseIsNull(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.m2o.domain.eq", []orm.FieldDefinition{
		{Name: "company_id", Type: orm.Many2One, Relation: "core.company"},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.m2o.domain.eq", [][]interface{}{
		{"company_id", "=", false},
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

func TestBuildSearchWhereClauseTextEqualsFalseIsEmpty(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.text.domain", []orm.FieldDefinition{
		{Name: "note", Type: orm.Text},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.text.domain", [][]interface{}{
		{"note", "=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if !strings.Contains(where, "IS NULL") || !strings.Contains(where, "= ''") {
		t.Fatalf("expected empty text clause, got %q", where)
	}
}

func TestBuildSearchWhereClauseSelectionNotFalseIsSet(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.selection.domain", []orm.FieldDefinition{
		{Name: "state", Type: orm.Selection},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.selection.domain", [][]interface{}{
		{"state", "!=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if !strings.Contains(where, "IS NOT NULL") || !strings.Contains(where, "<> ''") {
		t.Fatalf("expected set selection clause, got %q", where)
	}
}

func TestBuildSearchWhereClauseIntegerEqualsFalseIsNullOrZero(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.int.domain", []orm.FieldDefinition{
		{Name: "color", Type: orm.Integer},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.int.domain", [][]interface{}{
		{"color", "=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if !strings.Contains(where, "IS NULL") || !strings.Contains(where, "= 0") {
		t.Fatalf("expected null-or-zero clause, got %q", where)
	}
}

func TestBuildSearchWhereClauseOne2ManyNotFalseUsesExists(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.child.domain", []orm.FieldDefinition{
		{Name: "parent_id", Type: orm.Many2One, Relation: "test.parent.domain"},
	})
	orm.RegisterStubModelForTest(t, "test.parent.domain", []orm.FieldDefinition{
		{Name: "line_ids", Type: orm.One2Many, Relation: "test.child.domain"},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.parent.domain", [][]interface{}{
		{"line_ids", "!=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if !strings.Contains(where, "EXISTS") || !strings.Contains(where, "parent_id") {
		t.Fatalf("expected EXISTS on inverse many2one, got %q", where)
	}
}

func TestBuildSearchWhereClauseMany2ManyEqualsFalseUsesNotExists(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.m2m.domain", []orm.FieldDefinition{
		{Name: "tag_ids", Type: orm.Many2Many, Relation: "test.tag.domain", RelationTable: "test_m2m_rel", Column1: "parent_id", Column2: "tag_id"},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.m2m.domain", [][]interface{}{
		{"tag_ids", "=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if !strings.Contains(where, "NOT EXISTS") || !strings.Contains(where, "test_m2m_rel") {
		t.Fatalf("expected NOT EXISTS on rel table, got %q", where)
	}
}

func TestRecordMatchesDomainFalsyMany2One(t *testing.T) {
	model := "test.falsy.m2o"
	orm.RegisterStubModelForTest(t, model, []orm.FieldDefinition{
		{Name: "team_id", Type: orm.Many2One, Relation: "crm.team"},
	})

	if !orm.RecordMatchesDomainForTest(model, map[string]interface{}{"team_id": int64(3)}, [][]interface{}{{"team_id", "!=", false}}) {
		t.Fatal("expected set many2one to match != false")
	}
	if orm.RecordMatchesDomainForTest(model, map[string]interface{}{"team_id": nil}, [][]interface{}{{"team_id", "!=", false}}) {
		t.Fatal("expected unset many2one to fail != false")
	}
	if !orm.RecordMatchesDomainForTest(model, map[string]interface{}{"team_id": nil}, [][]interface{}{{"team_id", "=", false}}) {
		t.Fatal("expected unset many2one to match = false")
	}
}

func TestRecordMatchesDomainFalsyTextAndBoolean(t *testing.T) {
	model := "test.falsy.text"
	orm.RegisterStubModelForTest(t, model, []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "active", Type: orm.Boolean},
	})

	if !orm.RecordMatchesDomainForTest(model, map[string]interface{}{"name": ""}, [][]interface{}{{"name", "=", false}}) {
		t.Fatal("expected empty char to match = false")
	}
	if orm.RecordMatchesDomainForTest(model, map[string]interface{}{"active": false}, [][]interface{}{{"active", "!=", false}}) {
		t.Fatal("boolean != false should require true")
	}
	if !orm.RecordMatchesDomainForTest(model, map[string]interface{}{"active": true}, [][]interface{}{{"active", "!=", false}}) {
		t.Fatal("boolean != false should match true")
	}
}

func TestBuildSearchWhereClauseNonDateFalseUsesDistinctFrom(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.bool.domain", []orm.FieldDefinition{
		{Name: "active", Type: orm.Boolean},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.bool.domain", [][]interface{}{
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
	orCount, leaves, ok := orm.SplitDomainORPrefixForTest([][]interface{}{
		{"|"},
		{"name", "=", "a"},
		{"name", "=", "b"},
	})
	if !ok || orCount != 1 || len(leaves) != 2 {
		t.Fatalf("valid OR prefix: orCount=%d leaves=%d ok=%v", orCount, len(leaves), ok)
	}

	orCount, _, ok = orm.SplitDomainORPrefixForTest([][]interface{}{
		{"|"},
		{"|"},
		{"a", "=", 1},
		{"b", "=", 2},
		{"c", "=", 3},
	})
	if !ok || orCount != 2 {
		t.Fatalf("two OR markers: orCount=%d ok=%v", orCount, ok)
	}

	_, _, ok = orm.SplitDomainORPrefixForTest([][]interface{}{
		{"|"},
		{"name", "=", "a"},
	})
	if ok {
		t.Fatal("expected invalid shape for mismatched leaf count")
	}

	orCount, _, ok = orm.SplitDomainORPrefixForTest([][]interface{}{{"name", "=", "a"}})
	if !ok || orCount != 0 {
		t.Fatalf("non-OR domain: orCount=%d ok=%v", orCount, ok)
	}
}

func TestBuildSearchWhereClauseORPrefix(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.or.domain", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.or.domain", [][]interface{}{
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
	orm.RegisterStubModelForTest(t, "test.list.domain", []orm.FieldDefinition{
		{Name: "state", Type: orm.Selection},
	})

	where, args, err := orm.BuildSearchWhereClauseForTest("test.list.domain", [][]interface{}{
		{"state", "in", []interface{}{}},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause in: %v", err)
	}
	if where != "FALSE" || len(args) != 0 {
		t.Fatalf("empty IN: where=%q args=%v", where, args)
	}

	where, args, err = orm.BuildSearchWhereClauseForTest("test.list.domain", [][]interface{}{
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
	orm.RegisterStubModelForTest(t, "test.and.domain", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "active", Type: orm.Boolean},
	})

	where, args, err := orm.BuildAndWhereClausesForTest("test.and.domain", [][][]interface{}{
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
	if !orm.RecordMatchesDomainForTest("", map[string]interface{}{"name": "a"}, domain) {
		t.Fatal("expected match on first leaf")
	}
	if !orm.RecordMatchesDomainForTest("", map[string]interface{}{"name": "b"}, domain) {
		t.Fatal("expected match on second leaf")
	}
	if orm.RecordMatchesDomainForTest("", map[string]interface{}{"name": "c"}, domain) {
		t.Fatal("expected no match")
	}
}
